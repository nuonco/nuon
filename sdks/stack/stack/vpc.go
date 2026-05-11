package stack

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	// vpcCIDR matches the TF module (install-stacks/aws/modules/vpc). Why:
	// app-side IaC (helm charts, app modules) often pins network CIDRs that
	// reference this range; drifting from TF would silently break customers.
	vpcCIDR = "10.128.0.0/16"

	// runnerSubnetCIDR is a dedicated /24 for the runner ASG. It uses the
	// private route table (NAT egress) so the EC2 instance can reach the
	// init script and the runner API without a public IP.
	runnerSubnetCIDR = "10.128.2.0/24"
)

var (
	publicSubnetCIDRs  = []string{"10.128.0.0/24", "10.128.16.0/24"}
	privateSubnetCIDRs = []string{"10.128.1.0/24", "10.128.17.0/24"}
)

// tagSpec emits the common tag set every resource the sandbox terraform reads
// expects: install/org/app IDs (filter keys) + Name. Per-resource extras
// (visibility, network.nuon.co/domain, kubernetes.io/*) are appended.
func tagSpec(rt ec2types.ResourceType, name string, st *State, extra ...ec2types.Tag) ec2types.TagSpecification {
	tags := []ec2types.Tag{
		{Key: aws.String("Name"), Value: aws.String(name)},
		{Key: aws.String(installIDTagKey), Value: aws.String(st.InstallID)},
	}
	if st.OrgID != "" {
		tags = append(tags, ec2types.Tag{Key: aws.String("org.nuon.co/id"), Value: aws.String(st.OrgID)})
	}
	if st.AppID != "" {
		tags = append(tags, ec2types.Tag{Key: aws.String("app.nuon.co/id"), Value: aws.String(st.AppID)})
	}
	tags = append(tags, extra...)
	return ec2types.TagSpecification{ResourceType: rt, Tags: tags}
}

// subnetTags returns the tag set per subnet role. Matches the canonical CFN
// template (sandbox/exp/aws-cloudformation-templates/v0.1.4/vpc/eks/default/stack.yaml).
// Public/private subnets carry kubernetes.io/cluster + kubernetes.io/role for EKS
// LB subnet auto-discovery; runner subnet does not (it's not an EKS subnet).
func subnetTags(role string, clusterName string) []ec2types.Tag {
	switch role {
	case "public":
		return []ec2types.Tag{
			{Key: aws.String("network.nuon.co/domain"), Value: aws.String("public")},
			{Key: aws.String("visibility"), Value: aws.String("public")},
			{Key: aws.String("kubernetes.io/cluster/" + clusterName), Value: aws.String("shared")},
			{Key: aws.String("kubernetes.io/role/elb"), Value: aws.String("1")},
		}
	case "private":
		return []ec2types.Tag{
			{Key: aws.String("network.nuon.co/domain"), Value: aws.String("internal")},
			{Key: aws.String("visibility"), Value: aws.String("private")},
			{Key: aws.String("kubernetes.io/cluster/" + clusterName), Value: aws.String("shared")},
			{Key: aws.String("kubernetes.io/role/internal-elb"), Value: aws.String("1")},
		}
	case "runner":
		return []ec2types.Tag{
			{Key: aws.String("network.nuon.co/domain"), Value: aws.String("runner")},
			{Key: aws.String("visibility"), Value: aws.String("private")},
		}
	}
	return nil
}

// CreateVPC ensures the VPC and all dependent network resources exist. Safe to
// call repeatedly: each step is a discover-or-create, with AWS as the source of
// truth and the state file used only as a cache.
func CreateVPC(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	azs, err := pickAZs(ctx, c, 2)
	if err != nil {
		return err
	}

	if err := ensureVPC(ctx, log, c, st); err != nil {
		return err
	}
	if err := ensureIGW(ctx, log, c, st); err != nil {
		return err
	}
	if err := ensureSubnets(ctx, log, c, st, azs); err != nil {
		return err
	}
	if err := ensureNAT(ctx, log, c, st); err != nil {
		return err
	}
	if err := ensurePublicRouteTable(ctx, log, c, st); err != nil {
		return err
	}
	if err := ensurePrivateRouteTable(ctx, log, c, st); err != nil {
		return err
	}
	if err := ensureRunnerSubnet(ctx, log, c, st, azs); err != nil {
		return err
	}
	if err := ensureRunnerSecurityGroup(ctx, log, c, st); err != nil {
		return err
	}
	return nil
}

// ensureRunnerSubnet creates the dedicated runner subnet in AZ[0] and
// associates it with the private route table. Why: the runner has no public
// IP — it relies on NAT for outbound. Putting it on the private route table
// matches the TF module exactly.
func ensureRunnerSubnet(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State, azs []string) error {
	existing, err := findSubnetsForVPC(ctx, c, st.VPCID, st.InstallID)
	if err != nil {
		return err
	}
	if id, ok := existing[runnerSubnetCIDR]; ok {
		st.RunnerSubnetID = id
	} else {
		name := fmt.Sprintf("nuon-%s-runner", st.InstallID)
		out, err := c.CreateSubnet(ctx, &ec2.CreateSubnetInput{
			VpcId:             &st.VPCID,
			AvailabilityZone:  aws.String(azs[0]),
			CidrBlock:         aws.String(runnerSubnetCIDR),
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeSubnet, name, st, subnetTags("runner", st.ClusterName)...)},
		})
		if err != nil {
			return fmt.Errorf("create runner subnet: %w", err)
		}
		st.RunnerSubnetID = aws.ToString(out.Subnet.SubnetId)
		log.Info("created runner subnet", "id", st.RunnerSubnetID)
	}
	if st.PrivateRouteTbl != "" {
		if err := ensureRouteAssociations(ctx, c, st.PrivateRouteTbl, []string{st.RunnerSubnetID}); err != nil {
			return err
		}
	}
	return st.Save()
}

// ensureRunnerSecurityGroup provisions the runner SG with intra-SG ingress
// (so future co-located components can talk to it) and full egress. Mirrors
// install-stacks/aws/modules/vpc default SG.
func ensureRunnerSecurityGroup(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	if st.RunnerSecurityGroupID != "" {
		// Verify it still exists; otherwise re-resolve.
		out, err := c.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{GroupIds: []string{st.RunnerSecurityGroupID}})
		if err == nil && len(out.SecurityGroups) > 0 {
			return nil
		}
		if err != nil && !IsAWSErrCode(err, "InvalidGroup.NotFound") {
			return fmt.Errorf("describe runner sg: %w", err)
		}
		st.RunnerSecurityGroupID = ""
	}
	// Resolve by tag in case state was lost.
	d, err := c.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{st.VPCID}},
			{Name: aws.String("tag:" + installIDTagKey), Values: []string{st.InstallID}},
			{Name: aws.String("tag:Name"), Values: []string{"nuon-" + st.InstallID + "-runner"}},
		},
	})
	if err != nil {
		return fmt.Errorf("describe sg by tag: %w", err)
	}
	if len(d.SecurityGroups) > 0 {
		st.RunnerSecurityGroupID = aws.ToString(d.SecurityGroups[0].GroupId)
	} else {
		name := "nuon-" + st.InstallID + "-runner"
		cr, err := c.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
			VpcId:       &st.VPCID,
			GroupName:   aws.String(name),
			Description: aws.String("Nuon runner SG"),
			TagSpecifications: []ec2types.TagSpecification{tagSpec(
				ec2types.ResourceTypeSecurityGroup, name, st,
				ec2types.Tag{Key: aws.String("network.nuon.co/domain"), Value: aws.String("runner")},
			)},
		})
		if err != nil {
			return fmt.Errorf("create runner sg: %w", err)
		}
		st.RunnerSecurityGroupID = aws.ToString(cr.GroupId)
		log.Info("created runner SG", "id", st.RunnerSecurityGroupID)
	}
	// Intra-SG ingress on all ports. AuthorizeSecurityGroupIngress is not
	// idempotent — InvalidPermission.Duplicate is the "already there" signal.
	if _, err := c.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: &st.RunnerSecurityGroupID,
		IpPermissions: []ec2types.IpPermission{{
			IpProtocol: aws.String("-1"),
			UserIdGroupPairs: []ec2types.UserIdGroupPair{{
				GroupId: &st.RunnerSecurityGroupID,
			}},
		}},
	}); err != nil && !IsAWSErrCode(err, "InvalidPermission.Duplicate") {
		return fmt.Errorf("authorize ingress: %w", err)
	}
	// Egress 0.0.0.0/0 all ports. NB: AWS auto-creates a default allow-all
	// egress on SG creation; AuthorizeSecurityGroupEgress for the same rule
	// returns Duplicate. Keep this call so reconcile after a manual revoke
	// still re-establishes it.
	if _, err := c.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId: &st.RunnerSecurityGroupID,
		IpPermissions: []ec2types.IpPermission{{
			IpProtocol: aws.String("-1"),
			IpRanges:   []ec2types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
		}},
	}); err != nil && !IsAWSErrCode(err, "InvalidPermission.Duplicate") {
		return fmt.Errorf("authorize egress: %w", err)
	}
	return st.Save()
}

func ensureVPC(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	id, err := resolveByTag(ctx, c, ec2types.ResourceTypeVpc, st.VPCID, st.InstallID, func(id string) (bool, error) {
		out, err := c.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{VpcIds: []string{id}})
		if err != nil {
			if IsAWSErrCode(err, "InvalidVpcID.NotFound") {
				return false, nil
			}
			return false, err
		}
		return len(out.Vpcs) > 0, nil
	})
	if err != nil {
		return err
	}
	if id != "" {
		st.VPCID = id
	} else {
		out, err := c.CreateVpc(ctx, &ec2.CreateVpcInput{
			CidrBlock:         aws.String(vpcCIDR),
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeVpc, "nuon-"+st.InstallID, st)},
		})
		if err != nil {
			return fmt.Errorf("create vpc: %w", err)
		}
		st.VPCID = aws.ToString(out.Vpc.VpcId)
		log.Info("created VPC", "vpc_id", st.VPCID)
	}
	if err := reconcileVPCAttributes(ctx, c, st.VPCID); err != nil {
		return err
	}
	return st.Save()
}

func reconcileVPCAttributes(ctx context.Context, c *ec2.Client, vpcID string) error {
	out, err := c.DescribeVpcAttribute(ctx, &ec2.DescribeVpcAttributeInput{
		VpcId:     &vpcID,
		Attribute: ec2types.VpcAttributeNameEnableDnsHostnames,
	})
	if err != nil {
		return fmt.Errorf("describe vpc attribute: %w", err)
	}
	if out.EnableDnsHostnames != nil && aws.ToBool(out.EnableDnsHostnames.Value) {
		return nil
	}
	if _, err := c.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
		VpcId:              &vpcID,
		EnableDnsHostnames: &ec2types.AttributeBooleanValue{Value: aws.Bool(true)},
	}); err != nil {
		return fmt.Errorf("enable dns hostnames: %w", err)
	}
	return nil
}

func ensureIGW(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	id, err := resolveByTag(ctx, c, ec2types.ResourceTypeInternetGateway, st.IGWID, st.InstallID, func(id string) (bool, error) {
		out, err := c.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{InternetGatewayIds: []string{id}})
		if err != nil {
			if IsAWSErrCode(err, "InvalidInternetGatewayID.NotFound") {
				return false, nil
			}
			return false, err
		}
		return len(out.InternetGateways) > 0, nil
	})
	if err != nil {
		return err
	}
	if id == "" {
		out, err := c.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeInternetGateway, "nuon-"+st.InstallID, st)},
		})
		if err != nil {
			return fmt.Errorf("create igw: %w", err)
		}
		id = aws.ToString(out.InternetGateway.InternetGatewayId)
		log.Info("created IGW", "igw_id", id)
	}
	st.IGWID = id

	// Attach to VPC if not already attached. AttachInternetGateway is idempotent
	// for the same (igw, vpc) pair but errors on "already attached"; check first.
	out, err := c.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{InternetGatewayIds: []string{id}})
	if err != nil {
		return fmt.Errorf("describe igw: %w", err)
	}
	attached := false
	for _, g := range out.InternetGateways {
		for _, a := range g.Attachments {
			if aws.ToString(a.VpcId) == st.VPCID {
				attached = true
			}
		}
	}
	if !attached {
		if _, err := c.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
			InternetGatewayId: &id,
			VpcId:             &st.VPCID,
		}); err != nil {
			return fmt.Errorf("attach igw: %w", err)
		}
	}
	return st.Save()
}

func ensureSubnets(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State, azs []string) error {
	pubIDs, err := ensureSubnetSet(ctx, log, c, st, azs, publicSubnetCIDRs, true, "public", subnetTags("public", st.ClusterName))
	if err != nil {
		return err
	}
	privIDs, err := ensureSubnetSet(ctx, log, c, st, azs, privateSubnetCIDRs, false, "private", subnetTags("private", st.ClusterName))
	if err != nil {
		return err
	}
	st.PublicSubnetIDs = pubIDs
	st.PrivateSubnetIDs = privIDs
	return st.Save()
}

// ensureSubnetSet returns CIDR-ordered subnet IDs, creating any missing CIDRs.
// We index by CIDR (not list position) so partial state never causes us to
// create overlapping subnets.
func ensureSubnetSet(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State, azs []string, cidrs []string, public bool, label string, extraTags []ec2types.Tag) ([]string, error) {
	existing, err := findSubnetsForVPC(ctx, c, st.VPCID, st.InstallID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(cidrs))
	for i, cidr := range cidrs {
		if id, ok := existing[cidr]; ok {
			ids[i] = id
			continue
		}
		name := fmt.Sprintf("nuon-%s-%s-%d", st.InstallID, label, i)
		out, err := c.CreateSubnet(ctx, &ec2.CreateSubnetInput{
			VpcId:             &st.VPCID,
			AvailabilityZone:  aws.String(azs[i]),
			CidrBlock:         aws.String(cidr),
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeSubnet, name, st, extraTags...)},
		})
		if err != nil {
			return nil, fmt.Errorf("create subnet %s: %w", name, err)
		}
		ids[i] = aws.ToString(out.Subnet.SubnetId)
		log.Info("created subnet", "id", ids[i], "cidr", cidr, "label", label)
	}

	if public {
		for _, id := range ids {
			if err := reconcileSubnetPublicIP(ctx, c, id); err != nil {
				return nil, err
			}
		}
	}
	return ids, nil
}

// findSubnetsForVPC returns a CIDR→ID map of subnets in this install's VPC.
func findSubnetsForVPC(ctx context.Context, c *ec2.Client, vpcID, installID string) (map[string]string, error) {
	out, err := c.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			{Name: aws.String("tag:" + installIDTagKey), Values: []string{installID}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describe subnets: %w", err)
	}
	m := make(map[string]string, len(out.Subnets))
	for _, s := range out.Subnets {
		m[aws.ToString(s.CidrBlock)] = aws.ToString(s.SubnetId)
	}
	return m, nil
}

func reconcileSubnetPublicIP(ctx context.Context, c *ec2.Client, subnetID string) error {
	out, err := c.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{SubnetIds: []string{subnetID}})
	if err != nil {
		return fmt.Errorf("describe subnet: %w", err)
	}
	if len(out.Subnets) > 0 && aws.ToBool(out.Subnets[0].MapPublicIpOnLaunch) {
		return nil
	}
	if _, err := c.ModifySubnetAttribute(ctx, &ec2.ModifySubnetAttributeInput{
		SubnetId:            &subnetID,
		MapPublicIpOnLaunch: &ec2types.AttributeBooleanValue{Value: aws.Bool(true)},
	}); err != nil {
		return fmt.Errorf("public-ip-on-launch %s: %w", subnetID, err)
	}
	return nil
}

func ensureNAT(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	if err := ensureEIP(ctx, log, c, st); err != nil {
		return err
	}

	id, err := findActiveNAT(ctx, c, st.VPCID, st.InstallID)
	if err != nil {
		return err
	}
	if id == "" {
		out, err := c.CreateNatGateway(ctx, &ec2.CreateNatGatewayInput{
			AllocationId:      &st.NATElasticIP,
			SubnetId:          &st.PublicSubnetIDs[0],
			TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeNatgateway, "nuon-"+st.InstallID, st)},
		})
		if err != nil {
			return fmt.Errorf("create nat: %w", err)
		}
		id = aws.ToString(out.NatGateway.NatGatewayId)
		log.Info("created NAT gateway, waiting for AVAILABLE", "nat_id", id)
	}
	st.NATGatewayID = id
	if err := st.Save(); err != nil {
		return err
	}
	return waitNATAvailable(ctx, c, id)
}

// findActiveNAT returns a NAT in this VPC that's available or pending — i.e.
// usable on resume. Deleted/deleting/failed NATs are ignored so we'll create
// a fresh one.
func findActiveNAT(ctx context.Context, c *ec2.Client, vpcID, installID string) (string, error) {
	out, err := c.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{
		Filter: []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			{Name: aws.String("tag:" + installIDTagKey), Values: []string{installID}},
			{Name: aws.String("state"), Values: []string{"pending", "available"}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("describe nat: %w", err)
	}
	if len(out.NatGateways) == 0 {
		return "", nil
	}
	return aws.ToString(out.NatGateways[0].NatGatewayId), nil
}

func ensureEIP(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	out, err := c.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("tag:" + installIDTagKey), Values: []string{st.InstallID}},
		},
	})
	if err != nil {
		return fmt.Errorf("describe eip: %w", err)
	}
	if len(out.Addresses) > 0 {
		st.NATElasticIP = aws.ToString(out.Addresses[0].AllocationId)
		return st.Save()
	}
	alloc, err := c.AllocateAddress(ctx, &ec2.AllocateAddressInput{
		Domain:            ec2types.DomainTypeVpc,
		TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeElasticIp, "nuon-"+st.InstallID, st)},
	})
	if err != nil {
		return fmt.Errorf("allocate eip: %w", err)
	}
	st.NATElasticIP = aws.ToString(alloc.AllocationId)
	log.Info("allocated EIP", "alloc_id", st.NATElasticIP)
	return st.Save()
}

func ensurePublicRouteTable(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	rt, err := ensureRouteTable(ctx, log, c, st, "nuon-"+st.InstallID+"-public")
	if err != nil {
		return err
	}
	st.PublicRouteTable = rt
	if err := ensureRouteToIGW(ctx, c, rt, st.IGWID); err != nil {
		return err
	}
	if err := ensureRouteAssociations(ctx, c, rt, st.PublicSubnetIDs); err != nil {
		return err
	}
	return st.Save()
}

func ensurePrivateRouteTable(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State) error {
	rt, err := ensureRouteTable(ctx, log, c, st, "nuon-"+st.InstallID+"-private")
	if err != nil {
		return err
	}
	st.PrivateRouteTbl = rt
	if err := ensureRouteToNAT(ctx, c, rt, st.NATGatewayID); err != nil {
		return err
	}
	if err := ensureRouteAssociations(ctx, c, rt, st.PrivateSubnetIDs); err != nil {
		return err
	}
	return st.Save()
}

// ensureRouteTable resolves a route table by Name tag (one per role: public/
// private), creating it if missing. Name is the disambiguator since both route
// tables live in the same VPC under the same install_id tag.
func ensureRouteTable(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State, name string) (string, error) {
	out, err := c.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{st.VPCID}},
			{Name: aws.String("tag:" + installIDTagKey), Values: []string{st.InstallID}},
			{Name: aws.String("tag:Name"), Values: []string{name}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("describe route tables: %w", err)
	}
	if len(out.RouteTables) > 0 {
		return aws.ToString(out.RouteTables[0].RouteTableId), nil
	}
	cr, err := c.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
		VpcId:             &st.VPCID,
		TagSpecifications: []ec2types.TagSpecification{tagSpec(ec2types.ResourceTypeRouteTable, name, st)},
	})
	if err != nil {
		return "", fmt.Errorf("create route table %s: %w", name, err)
	}
	id := aws.ToString(cr.RouteTable.RouteTableId)
	log.Info("created route table", "id", id, "name", name)
	return id, nil
}

func ensureRouteToIGW(ctx context.Context, c *ec2.Client, rtID, igwID string) error {
	if has, err := hasDefaultRoute(ctx, c, rtID); err != nil || has {
		return err
	}
	_, err := c.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:         &rtID,
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            &igwID,
	})
	if err != nil && !IsAWSErrCode(err, "RouteAlreadyExists") {
		return fmt.Errorf("create public route: %w", err)
	}
	return nil
}

func ensureRouteToNAT(ctx context.Context, c *ec2.Client, rtID, natID string) error {
	if has, err := hasDefaultRoute(ctx, c, rtID); err != nil || has {
		return err
	}
	_, err := c.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:         &rtID,
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		NatGatewayId:         &natID,
	})
	if err != nil && !IsAWSErrCode(err, "RouteAlreadyExists") {
		return fmt.Errorf("create private route: %w", err)
	}
	return nil
}

func hasDefaultRoute(ctx context.Context, c *ec2.Client, rtID string) (bool, error) {
	out, err := c.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{RouteTableIds: []string{rtID}})
	if err != nil {
		return false, fmt.Errorf("describe route table: %w", err)
	}
	for _, t := range out.RouteTables {
		for _, r := range t.Routes {
			if aws.ToString(r.DestinationCidrBlock) == "0.0.0.0/0" {
				return true, nil
			}
		}
	}
	return false, nil
}

func ensureRouteAssociations(ctx context.Context, c *ec2.Client, rtID string, subnetIDs []string) error {
	out, err := c.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{RouteTableIds: []string{rtID}})
	if err != nil {
		return fmt.Errorf("describe route table assoc: %w", err)
	}
	already := map[string]bool{}
	for _, t := range out.RouteTables {
		for _, a := range t.Associations {
			if a.SubnetId != nil {
				already[aws.ToString(a.SubnetId)] = true
			}
		}
	}
	for _, sn := range subnetIDs {
		if already[sn] {
			continue
		}
		sn := sn
		if _, err := c.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{RouteTableId: &rtID, SubnetId: &sn}); err != nil {
			return fmt.Errorf("assoc rt: %w", err)
		}
	}
	return nil
}

// resolveByTag resolves an EC2 resource using cached state first, then a tag
// lookup, then returns "" if neither finds it (caller should create).
//
// verify must return (found, err) given a candidate ID.
func resolveByTag(ctx context.Context, c *ec2.Client, rt ec2types.ResourceType, cachedID, installID string, verify func(id string) (bool, error)) (string, error) {
	if cachedID != "" {
		ok, err := verify(cachedID)
		if err != nil {
			return "", err
		}
		if ok {
			return cachedID, nil
		}
	}
	ids, err := findEC2ResourcesByInstallID(ctx, c, rt, installID)
	if err != nil {
		return "", err
	}
	for _, id := range ids {
		if id == cachedID {
			continue // already checked above
		}
		ok, err := verify(id)
		if err != nil {
			return "", err
		}
		if ok {
			return id, nil
		}
	}
	return "", nil
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
	slices.Sort(azs)
	return azs, nil
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
	if st.RunnerSecurityGroupID != "" {
		if _, err := c.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{GroupId: &st.RunnerSecurityGroupID}); err != nil {
			log.Warn("delete runner sg", "id", st.RunnerSecurityGroupID, "err", err)
		}
		st.RunnerSecurityGroupID = ""
	}
	if st.RunnerSubnetID != "" {
		if _, err := c.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{SubnetId: &st.RunnerSubnetID}); err != nil {
			log.Warn("delete runner subnet", "id", st.RunnerSubnetID, "err", err)
		}
		st.RunnerSubnetID = ""
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
