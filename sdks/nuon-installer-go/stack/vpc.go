package stack

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	vpcCIDR = "10.42.0.0/16"
)

var (
	publicSubnetCIDRs  = []string{"10.42.0.0/20", "10.42.16.0/20", "10.42.32.0/20"}
	privateSubnetCIDRs = []string{"10.42.64.0/20", "10.42.80.0/20", "10.42.96.0/20"}
)

func tagSpec(rt ec2types.ResourceType, name, installID string) ec2types.TagSpecification {
	return ec2types.TagSpecification{
		ResourceType: rt,
		Tags: []ec2types.Tag{
			{Key: aws.String("Name"), Value: aws.String(name)},
			{Key: aws.String("nuon:install_id"), Value: aws.String(installID)},
		},
	}
}

// CreateVPC builds a minimal 3-AZ VPC: VPC, IGW, 3 public + 3 private subnets,
// single NAT gateway in the first public subnet, and route tables.
func CreateVPC(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	azs, err := pickAZs(ctx, c, 3)
	if err != nil {
		return err
	}

	if st.VPCID == "" {
		out, err := c.CreateVpc(ctx, &ec2.CreateVpcInput{
			CidrBlock:         aws.String(vpcCIDR),
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeVpc, "nuon-"+st.InstallID, st.InstallID)},
		})
		if err != nil {
			return fmt.Errorf("create vpc: %w", err)
		}
		st.VPCID = aws.ToString(out.Vpc.VpcId)
		log.Info("created VPC", "vpc_id", st.VPCID)
		if _, err := c.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
			VpcId:              &st.VPCID,
			EnableDnsHostnames: &ec2types.AttributeBooleanValue{Value: aws.Bool(true)},
		}); err != nil {
			return fmt.Errorf("enable dns hostnames: %w", err)
		}
		if err := st.Save(); err != nil {
			return err
		}
	}

	if st.IGWID == "" {
		out, err := c.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeInternetGateway, "nuon-"+st.InstallID, st.InstallID)},
		})
		if err != nil {
			return fmt.Errorf("create igw: %w", err)
		}
		st.IGWID = aws.ToString(out.InternetGateway.InternetGatewayId)
		if _, err := c.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
			InternetGatewayId: &st.IGWID,
			VpcId:             &st.VPCID,
		}); err != nil {
			return fmt.Errorf("attach igw: %w", err)
		}
		log.Info("created IGW", "igw_id", st.IGWID)
		if err := st.Save(); err != nil {
			return err
		}
	}

	if len(st.PublicSubnetIDs) < len(publicSubnetCIDRs) {
		st.PublicSubnetIDs = nil
		for i, cidr := range publicSubnetCIDRs {
			id, err := createSubnet(ctx, c, st, azs[i], cidr, true, fmt.Sprintf("nuon-%s-public-%d", st.InstallID, i))
			if err != nil {
				return err
			}
			st.PublicSubnetIDs = append(st.PublicSubnetIDs, id)
		}
		log.Info("created public subnets", "ids", st.PublicSubnetIDs)
		if err := st.Save(); err != nil {
			return err
		}
	}
	if len(st.PrivateSubnetIDs) < len(privateSubnetCIDRs) {
		st.PrivateSubnetIDs = nil
		for i, cidr := range privateSubnetCIDRs {
			id, err := createSubnet(ctx, c, st, azs[i], cidr, false, fmt.Sprintf("nuon-%s-private-%d", st.InstallID, i))
			if err != nil {
				return err
			}
			st.PrivateSubnetIDs = append(st.PrivateSubnetIDs, id)
		}
		log.Info("created private subnets", "ids", st.PrivateSubnetIDs)
		if err := st.Save(); err != nil {
			return err
		}
	}

	if st.NATElasticIP == "" {
		out, err := c.AllocateAddress(ctx, &ec2.AllocateAddressInput{
			Domain:            ec2types.DomainTypeVpc,
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeElasticIp, "nuon-"+st.InstallID, st.InstallID)},
		})
		if err != nil {
			return fmt.Errorf("allocate eip: %w", err)
		}
		st.NATElasticIP = aws.ToString(out.AllocationId)
		if err := st.Save(); err != nil {
			return err
		}
	}
	if st.NATGatewayID == "" {
		out, err := c.CreateNatGateway(ctx, &ec2.CreateNatGatewayInput{
			AllocationId:      &st.NATElasticIP,
			SubnetId:          &st.PublicSubnetIDs[0],
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeNatgateway, "nuon-"+st.InstallID, st.InstallID)},
		})
		if err != nil {
			return fmt.Errorf("create nat: %w", err)
		}
		st.NATGatewayID = aws.ToString(out.NatGateway.NatGatewayId)
		log.Info("created NAT gateway, waiting for AVAILABLE", "nat_id", st.NATGatewayID)
		if err := st.Save(); err != nil {
			return err
		}
		if err := waitNATAvailable(ctx, c, st.NATGatewayID); err != nil {
			return err
		}
	}

	if st.PublicRouteTable == "" {
		rt, err := createRouteTable(ctx, c, st, "nuon-"+st.InstallID+"-public")
		if err != nil {
			return err
		}
		if _, err := c.CreateRoute(ctx, &ec2.CreateRouteInput{
			RouteTableId:         &rt,
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
			GatewayId:            &st.IGWID,
		}); err != nil {
			return fmt.Errorf("create public route: %w", err)
		}
		for _, sn := range st.PublicSubnetIDs {
			sn := sn
			if _, err := c.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{RouteTableId: &rt, SubnetId: &sn}); err != nil {
				return fmt.Errorf("assoc public rt: %w", err)
			}
		}
		st.PublicRouteTable = rt
		if err := st.Save(); err != nil {
			return err
		}
	}
	if st.PrivateRouteTbl == "" {
		rt, err := createRouteTable(ctx, c, st, "nuon-"+st.InstallID+"-private")
		if err != nil {
			return err
		}
		if _, err := c.CreateRoute(ctx, &ec2.CreateRouteInput{
			RouteTableId:         &rt,
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
			NatGatewayId:         &st.NATGatewayID,
		}); err != nil {
			return fmt.Errorf("create private route: %w", err)
		}
		for _, sn := range st.PrivateSubnetIDs {
			sn := sn
			if _, err := c.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{RouteTableId: &rt, SubnetId: &sn}); err != nil {
				return fmt.Errorf("assoc private rt: %w", err)
			}
		}
		st.PrivateRouteTbl = rt
		if err := st.Save(); err != nil {
			return err
		}
	}

	return nil
}

func pickAZs(ctx context.Context, c *ec2.Client, n int) ([]string, error) {
	out, err := c.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return nil, fmt.Errorf("describe AZs: %w", err)
	}
	if len(out.AvailabilityZones) < n {
		return nil, fmt.Errorf("need %d AZs, have %d", n, len(out.AvailabilityZones))
	}
	azs := make([]string, 0, n)
	for _, z := range out.AvailabilityZones[:n] {
		azs = append(azs, aws.ToString(z.ZoneName))
	}
	return azs, nil
}

func createSubnet(ctx context.Context, c *ec2.Client, st *State, az, cidr string, public bool, name string) (string, error) {
	out, err := c.CreateSubnet(ctx, &ec2.CreateSubnetInput{
		VpcId:             &st.VPCID,
		AvailabilityZone:  &az,
		CidrBlock:         &cidr,
		TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeSubnet, name, st.InstallID)},
	})
	if err != nil {
		return "", fmt.Errorf("create subnet %s: %w", name, err)
	}
	id := aws.ToString(out.Subnet.SubnetId)
	if public {
		if _, err := c.ModifySubnetAttribute(ctx, &ec2.ModifySubnetAttributeInput{
			SubnetId:            &id,
			MapPublicIpOnLaunch: &ec2types.AttributeBooleanValue{Value: aws.Bool(true)},
		}); err != nil {
			return "", fmt.Errorf("public-ip-on-launch %s: %w", name, err)
		}
	}
	return id, nil
}

func createRouteTable(ctx context.Context, c *ec2.Client, st *State, name string) (string, error) {
	out, err := c.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
		VpcId:             &st.VPCID,
		TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeRouteTable, name, st.InstallID)},
	})
	if err != nil {
		return "", fmt.Errorf("create route table %s: %w", name, err)
	}
	return aws.ToString(out.RouteTable.RouteTableId), nil
}

func waitNATAvailable(ctx context.Context, c *ec2.Client, id string) error {
	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		out, err := c.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{NatGatewayIds: []string{id}})
		if err != nil {
			return err
		}
		if len(out.NatGateways) > 0 && out.NatGateways[0].State == ec2types.NatGatewayStateAvailable {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	return fmt.Errorf("nat gateway %s not available within timeout", id)
}

// DeleteVPC tears down the VPC and dependent resources in reverse order. Idempotent.
func DeleteVPC(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	if st.NATGatewayID != "" {
		if _, err := c.DeleteNatGateway(ctx, &ec2.DeleteNatGatewayInput{NatGatewayId: &st.NATGatewayID}); err != nil {
			log.Warn("delete nat (continuing)", "err", err)
		} else {
			log.Info("deleting NAT gateway", "nat_id", st.NATGatewayID)
		}
		// NAT delete is async; wait for DELETED
		deadline := time.Now().Add(10 * time.Minute)
		for time.Now().Before(deadline) {
			out, err := c.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{NatGatewayIds: []string{st.NATGatewayID}})
			if err != nil || len(out.NatGateways) == 0 {
				break
			}
			if out.NatGateways[0].State == ec2types.NatGatewayStateDeleted {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Second):
			}
		}
		st.NATGatewayID = ""
	}
	if st.NATElasticIP != "" {
		if _, err := c.ReleaseAddress(ctx, &ec2.ReleaseAddressInput{AllocationId: &st.NATElasticIP}); err != nil {
			log.Warn("release eip", "err", err)
		}
		st.NATElasticIP = ""
	}
	for _, rt := range []*string{&st.PublicRouteTable, &st.PrivateRouteTbl} {
		if *rt == "" {
			continue
		}
		// Disassociate first
		out, err := c.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{RouteTableIds: []string{*rt}})
		if err == nil {
			for _, t := range out.RouteTables {
				for _, a := range t.Associations {
					if aws.ToBool(a.Main) {
						continue
					}
					_, _ = c.DisassociateRouteTable(ctx, &ec2.DisassociateRouteTableInput{AssociationId: a.RouteTableAssociationId})
				}
			}
		}
		if _, err := c.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{RouteTableId: rt}); err != nil {
			log.Warn("delete route table", "rt", *rt, "err", err)
		}
		*rt = ""
	}
	for _, ids := range [][]string{st.PublicSubnetIDs, st.PrivateSubnetIDs} {
		for _, id := range ids {
			id := id
			if _, err := c.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{SubnetId: &id}); err != nil {
				log.Warn("delete subnet", "id", id, "err", err)
			}
		}
	}
	st.PublicSubnetIDs = nil
	st.PrivateSubnetIDs = nil
	if st.IGWID != "" {
		if _, err := c.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{InternetGatewayId: &st.IGWID, VpcId: &st.VPCID}); err != nil {
			log.Warn("detach igw", "err", err)
		}
		if _, err := c.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{InternetGatewayId: &st.IGWID}); err != nil {
			log.Warn("delete igw", "err", err)
		}
		st.IGWID = ""
	}
	if st.VPCID != "" {
		if _, err := c.DeleteVpc(ctx, &ec2.DeleteVpcInput{VpcId: &st.VPCID}); err != nil {
			log.Warn("delete vpc", "err", err)
		} else {
			log.Info("deleted VPC", "vpc_id", st.VPCID)
		}
		st.VPCID = ""
	}
	return st.Save()
}
