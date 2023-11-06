CREATE TABLE IF NOT EXISTS trivysbom (
	id                    UUID,
	schema                String,
	bom_format            String,
	spec_version          String,
	serial_number         String,
	version               INTEGER,
	metadata_timestamp    DateTime('UTC'),
	metatool_vendor       String,
	metatool_name         String,
	metatool_version      String,
	component_bom_ref     String,
	component_type        String,
	component_name        String,
	component_version     String,
	component_property_name String,
	component_property_value String,
	component_hash_alg    String,
	component_hash_content String,
	component_license_exp String,
	component_purl        String,
	dependency_ref        String,
	vulnerabilities 	String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
