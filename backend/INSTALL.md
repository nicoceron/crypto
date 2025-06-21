# Prerequisites Installation Guide

Before deploying your stock analyzer to AWS Lambda, you need these tools installed:

## 🔧 Required Tools

### 1. Go (Already Installed ✅)

Your Go installation is working correctly.

### 2. Terraform (Already Installed ✅)

Your Terraform installation is working correctly.

### 3. AWS CLI (❌ Missing)

**macOS (Homebrew):**

```bash
brew install awscli
```

**macOS (Official Installer):**

```bash
curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
sudo installer -pkg AWSCLIV2.pkg -target /
```

**Linux:**

```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

**Windows:**
Download and run: https://awscli.amazonaws.com/AWSCLIV2.msi

### 4. jq (Already Installed ✅)

Your jq installation is working correctly.

## 🔑 AWS Configuration

After installing AWS CLI, configure your credentials:

```bash
aws configure
```

You'll need:

- **AWS Access Key ID**: From your AWS account
- **AWS Secret Access Key**: From your AWS account
- **Default region**: Recommend `us-west-2`
- **Default output format**: Recommend `json`

### Getting AWS Credentials

1. **Log into AWS Console**: https://console.aws.amazon.com
2. **Go to IAM** → Users → Your user → Security credentials
3. **Create Access Key** → Command Line Interface (CLI)
4. **Download** the CSV or copy the credentials

### Required AWS Permissions

Your AWS user needs these permissions:

- EC2 (VPC, subnets, security groups)
- Lambda (functions, permissions)
- API Gateway (REST APIs)
- RDS (PostgreSQL instances)
- S3 (deployment buckets)
- IAM (roles, policies)
- CloudWatch (logs, monitoring)
- Secrets Manager (database credentials)

For development, you can use the `PowerUserAccess` managed policy.

## 🚀 Quick Start

After installing prerequisites:

```bash
# Test AWS connection
aws sts get-caller-identity

# If successful, run setup
./scripts/setup.sh
```

## 🔍 Troubleshooting

### AWS CLI Not Found

```bash
# Check if installed
which aws

# If not found, check PATH
echo $PATH

# Restart terminal after installation
```

### AWS Credentials Invalid

```bash
# Re-configure
aws configure

# Test connection
aws sts get-caller-identity
```

### Permission Denied

```bash
# Make scripts executable
chmod +x scripts/*.sh
```

## 💡 Tips

- Use AWS profiles for multiple accounts: `aws configure --profile myprofile`
- Keep credentials secure and never commit them to git
- Consider using AWS SSO for enterprise environments
- Use least privilege principle for production deployments

---

Once all prerequisites are installed, proceed with: `./scripts/setup.sh`
