# ---------------------------------------------------------------
# karpenter-image-cache
#
# Creates an EBS snapshot containing pre-pulled container images.
# Karpenter EC2NodeClass can mount this snapshot as a secondary
# volume so nodes start with images already in containerd's
# content store — avoiding multi-GB pulls on every scale-out.
#
# Flow:
#   1. Launch a temporary EC2 builder instance
#   2. Attach a secondary EBS volume
#   3. Pull all specified images into containerd on that volume
#   4. Instance shuts itself down after pulling
#   5. Snapshot the volume
#   6. Output snapshot_id for use in EC2NodeClass blockDeviceMappings
# ---------------------------------------------------------------

locals {
  name_prefix = "image-cache"
  images_hash = md5(join(",", sort(var.images)))
}

# ---------------------------------------------------------------
# Data sources
# ---------------------------------------------------------------

data "aws_subnet" "selected" {
  id = var.subnet_id
}

data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "state"
    values = ["available"]
  }
}

# ---------------------------------------------------------------
# IAM – builder instance profile with ECR read + SSM access
# ---------------------------------------------------------------

resource "aws_iam_role" "builder" {
  name_prefix = "${local.name_prefix}-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.builder.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.builder.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "builder" {
  name_prefix = "${local.name_prefix}-"
  role        = aws_iam_role.builder.name
  tags        = var.tags
}

# ---------------------------------------------------------------
# Security group – outbound-only (no inbound / no SSH)
# ---------------------------------------------------------------

resource "aws_security_group" "builder" {
  name_prefix = "${local.name_prefix}-"
  description = "Image cache builder - egress only for pulling container images"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound for image pulls"
  }

  tags = merge(var.tags, { Name = "${local.name_prefix}-builder" })
}

# ---------------------------------------------------------------
# Cache EBS volume
# ---------------------------------------------------------------

resource "aws_ebs_volume" "cache" {
  availability_zone = data.aws_subnet.selected.availability_zone
  size              = var.volume_size_gb
  type              = var.volume_type

  tags = merge(var.tags, {
    Name        = "${local.name_prefix}"
    images_hash = local.images_hash
  })
}

# ---------------------------------------------------------------
# Builder instance
#
# user_data formats the volume, pulls images via containerd,
# unmounts, and shuts down. The instance stays in "stopped"
# state until terraform destroy terminates it.
# ---------------------------------------------------------------

resource "aws_instance" "builder" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  iam_instance_profile   = aws_iam_instance_profile.builder.name
  vpc_security_group_ids = [aws_security_group.builder.id]

  user_data = templatefile("${path.module}/scripts/pull-images.sh.tftpl", {
    images = var.images
  })

  # Small root volume — all image data goes to the secondary cache volume
  root_block_device {
    volume_size           = 20
    volume_type           = "gp3"
    delete_on_termination = true
  }

  tags = merge(var.tags, {
    Name        = "${local.name_prefix}-builder"
    images_hash = local.images_hash
  })

  # Recreate when images change
  lifecycle {
    replace_triggered_by = [aws_ebs_volume.cache]
  }
}

# ---------------------------------------------------------------
# Attach the cache volume to the builder
# ---------------------------------------------------------------

resource "aws_volume_attachment" "cache" {
  device_name = "/dev/xvdf"
  volume_id   = aws_ebs_volume.cache.id
  instance_id = aws_instance.builder.id

  # Don't try to detach on destroy — instance termination handles it
  skip_destroy = true
}

# ---------------------------------------------------------------
# Wait for the builder to finish (it shuts down when done)
# ---------------------------------------------------------------

resource "null_resource" "wait_for_completion" {
  depends_on = [aws_volume_attachment.cache]

  triggers = {
    instance_id = aws_instance.builder.id
    images_hash = local.images_hash
  }

  provisioner "local-exec" {
    command = <<-SCRIPT
      echo "Waiting for image cache builder ${aws_instance.builder.id} to complete..."
      while true; do
        STATE=$(aws ec2 describe-instances \
          --instance-ids ${aws_instance.builder.id} \
          --query 'Reservations[0].Instances[0].State.Name' \
          --output text \
          --region ${var.region} 2>/dev/null)
        echo "  instance state: $STATE"
        case "$STATE" in
          stopped)
            echo "Builder completed successfully."
            break
            ;;
          terminated|shutting-down)
            echo "ERROR: Builder instance terminated unexpectedly."
            exit 1
            ;;
        esac
        sleep 15
      done
    SCRIPT
  }
}

# ---------------------------------------------------------------
# Snapshot the cache volume
# ---------------------------------------------------------------

resource "aws_ebs_snapshot" "cache" {
  volume_id   = aws_ebs_volume.cache.id
  description = "Container image cache (${length(var.images)} images)"

  tags = merge(var.tags, {
    Name        = local.name_prefix
    images_hash = local.images_hash
  })

  depends_on = [null_resource.wait_for_completion]
}
