terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = ">= 1.0.0"
    }
  }
}

variable "linode_token" {
  description = "Linode API Token"
  type        = string
  sensitive   = true
}

provider "linode" {
  token = var.linode_token
}

data "http" "ipv6_addr" {
  url = "https://api64.ipify.org?format=text"
}

data "http" "ipv4_addr" {
  url = "https://api.ipify.org?format=text"
}

locals {
  valid_ipv4_pattern = "^(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)$$"
  valid_ipv6_pattern = "^(?:[a-fA-F0-9]{1,4}:){7}[a-fA-F0-9]{1,4}|(?:[a-fA-F0-9]{1,4}:){1,7}:|(?:[a-fA-F0-9]{1,4}:){1,6}:[a-fA-F0-9]{1,4}|(?:[a-fA-F0-9]{1,4}:){1,5}:(?:[a-fA-F0-9]{1,4}:){1,4}|(?:[a-fA-F0-9]{1,4}:){1,4}:(?:[a-fA-F0-9]{1,4}:){1,4}|(?:[a-fA-F0-9]{1,4}:){1,3}:(?:[a-fA-F0-9]{1,4}:){1,4}|(?:[a-fA-F0-9]{1,4}:){1,2}:(?:[a-fA-F0-9]{1,4}:){1,4}|[a-fA-F0-9]{1,4}:(?:[a-fA-F0-9]{1,4}:){1,4}|(?:[a-fA-F0-9]{1,4}:){1,4}:|:(?::[a-fA-F0-9]{1,4}){1,7}|::|(?:[a-fA-F0-9]{1,4}:){1,7}(?::[a-fA-F0-9]{1,4}){1,7}$$"
  valid_ipv4 = can(regex(local.valid_ipv4_pattern, data.http.ipv4_addr.response_body))
  valid_ipv6 = can(regex(local.valid_ipv6_pattern, data.http.ipv6_addr.response_body))
  ipv4_address = local.valid_ipv4 ? "${data.http.ipv4_addr.response_body}/32" : null
  ipv6_address = local.valid_ipv6 ? "${data.http.ipv6_addr.response_body}/128" : null
}

resource "linode_firewall" "cloud_linode_firewall" {
  label = "linode_cloud_firewall"

  outbound_policy = "ACCEPT"
  inbound_policy  = "DROP"

  dynamic "inbound" {
    for_each = local.valid_ipv4 || local.valid_ipv6 ? [1] : []
    content {
      label    = "tcp_inbound_ssh_accept_local"
      action   = "ACCEPT"
      ipv4     = local.ipv4_address != null ? [local.ipv4_address] : []
      ipv6     = local.ipv6_address != null ? [local.ipv6_address] : []
      protocol = "TCP"
      ports    = "22"
    }
  }
}

resource "null_resource" "validation_check" {
  count = local.valid_ipv4 || local.valid_ipv6 ? 0 : 1

  provisioner "local-exec" {
    command = "echo 'Both IPv4 and IPv6 addresses are invalid. Creating default cloud firewall without any rule.'"
  }
}
