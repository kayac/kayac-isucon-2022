packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "isucon" {
  ami_name      = "kayac-isucon-2022-${formatdate("YYYYMMDD-hhmm", timestamp())}"
  instance_type = "t3.medium"
  region        = "ap-northeast-1"

  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-focal-20.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }

  ssh_username = "ubuntu"
  encrypt_boot = false

  launch_block_device_mappings {
    device_name = "/dev/sda1"
    volume_type = "gp3"
    volume_size = 16
    delete_on_termination = true
  }
}

build {
  name = "isucon"
  sources = [
    "source.amazon-ebs.isucon"
  ]

  provisioner "file" {
    source      = "isucon.tar.gz"
    destination = "/tmp/isucon.tar.gz"
  }

  provisioner "file" {
    source      = "provisioning.sh"
    destination = "/tmp/provisioning.sh"
  }

  provisioner "shell" {
    environment_vars = []
    inline = [
      "sleep 10",
      "sudo /tmp/provisioning.sh",
      "sudo rm -fr /tmp/isucon* /tmp/provisioning.sh",
    ]
  }
}
