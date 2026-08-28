# Import using `${IP_POOL_ID}/${RANGE_ID}`.
terraform import oxide_system_ip_pool_range.example 4f0e69ad-66b6-41c0-b727-7b0285b0c384/3e2c6e84-bed8-4c94-afc3-1032082d6a90

# Alternatively, import using `${IP_POOL_ID}/${FIRST_ADDRESS}/${LAST_ADDRESS}`.
terraform import oxide_system_ip_pool_range.example 4f0e69ad-66b6-41c0-b727-7b0285b0c384/172.20.18.227/172.20.18.239
