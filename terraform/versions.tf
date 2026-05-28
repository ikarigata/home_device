terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
  }

  # state は当面ローカル。複数端末/チームで運用するなら下記のように S3 backend へ移行する。
  # backend "s3" {
  #   bucket = "<tfstate-bucket>"
  #   key    = "home_device/terraform.tfstate"
  #   region = "ap-northeast-1"
  # }
}
