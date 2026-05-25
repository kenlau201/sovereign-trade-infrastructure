environment            = "production"
aws_region             = "us-east-1"
project_name           = "sovereign-trade"
vpc_cidr               = "10.0.0.0/16"
availability_zones     = ["us-east-1a", "us-east-1b", "us-east-1c"]
eks_cluster_version    = "1.27"
eks_desired_capacity   = 5
eks_min_capacity       = 3
eks_max_capacity       = 10
instance_type          = "t3.2xlarge"
rds_instance_class     = "db.r6g.xlarge"
rds_allocated_storage  = 500
redis_node_type        = "cache.r6g.xlarge"
redis_num_cache_nodes  = 3

tags = {
  Project     = "sovereign-trade"
  Environment = "production"
  ManagedBy   = "terraform"
}
