environment            = "staging"
aws_region             = "us-east-1"
project_name           = "sovereign-trade"
vpc_cidr               = "10.0.0.0/16"
availability_zones     = ["us-east-1a", "us-east-1b", "us-east-1c"]
eks_cluster_version    = "1.27"
eks_desired_capacity   = 2
eks_min_capacity       = 1
eks_max_capacity       = 5
instance_type          = "t3.xlarge"
rds_instance_class     = "db.t3.large"
rds_allocated_storage  = 100
redis_node_type        = "cache.t3.medium"
redis_num_cache_nodes  = 2

tags = {
  Project     = "sovereign-trade"
  Environment = "staging"
  ManagedBy   = "terraform"
}
