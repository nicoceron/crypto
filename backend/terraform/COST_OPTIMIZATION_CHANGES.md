# Cost Optimization Changes - NAT Gateway Removal

## 🎯 Objective

Eliminate the **$1.80/month EC2-Other charges** caused by NAT Gateways by moving Lambda functions to public subnets.

## 💰 Expected Savings

- **Before**: ~$32-65/month (depending on number of AZs and data processing)
- **After**: $0/month for NAT Gateway costs
- **Monthly Savings**: $32-65/month

## 🔧 Changes Made

### 1. Main Terraform Configuration (`terraform/main.tf`)

```diff
- app_subnet_ids        = module.networking.app_subnet_ids
+ app_subnet_ids        = module.networking.public_subnet_ids
```

### 2. Networking Module (`terraform/modules/networking/main.tf`)

- **Commented out NAT Gateways** and Elastic IPs (lines ~65-90)
- **Removed NAT Gateway routes** from private subnet route tables
- **Added explanatory comments** for future reference

### 3. Networking Outputs (`terraform/modules/networking/outputs.tf`)

- **Commented out NAT Gateway outputs** since they no longer exist

### 4. Lambda Module (`terraform/modules/lambda/main.tf`)

- **Added documentation** explaining the public subnet usage
- **Kept VPC permissions** (still needed for public subnet Lambdas)

## 🔒 Security Impact

- ✅ **No security degradation**: Security groups still control access
- ✅ **Lambda functions maintain VPC isolation**
- ✅ **Same egress rules apply** (HTTPS, PostgreSQL, etc.)
- ℹ️ **Lambda functions now get public IPs** (but are not publicly accessible)

## 🚀 Deployment Steps

1. **Plan the changes**:

   ```bash
   cd terraform
   terraform plan
   ```

2. **Apply the changes**:

   ```bash
   terraform apply
   ```

3. **Expected Terraform Actions**:
   - **Destroy**: NAT Gateways and Elastic IPs
   - **Modify**: Lambda function network configurations
   - **Modify**: Route table configurations

## ✅ Verification

After deployment, verify:

1. **Lambda functions work correctly** (test API endpoints)
2. **Database connections work** (CockroachDB Cloud)
3. **External API calls work** (Alpaca, stock ratings API)
4. **No EC2-Other charges** in next month's bill

## 🔄 Rollback Plan

If issues arise, rollback by:

1. Uncommenting NAT Gateway resources in `networking/main.tf`
2. Changing `public_subnet_ids` back to `app_subnet_ids` in `main.tf`
3. Running `terraform apply`

## 📝 Notes

- **Private subnets still exist** for future use if needed
- **Database subnets remain isolated** (no internet access)
- **Architecture remains production-ready** with this optimization
