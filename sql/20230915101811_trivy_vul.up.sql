CREATE TABLE IF NOT EXISTS trivy_vul (
	id                    UUID,
	cluster_name          String,
	namespace             String,
	kind                  String,
	name                  String,
	vul_id                String,
	vul_vendor_ids        String,
	vul_pkg_id            String,
	vul_pkg_name          String,
	vul_pkg_path          String,
	vul_installed_version String,
	vul_fixed_version     String,
	vul_title             String,
	vul_severity          String,
	vul_published_date    DateTime('UTC'),
	vul_last_modified_date DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

