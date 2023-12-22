CREATE TABLE IF NOT EXISTS trivysbom (
	id                    UUID,
	cluster_name String,
	image_name            String,
	package_name String,
	package_url           String,
	bom_ref 			  String,
	serial_number         String,
	version 			  INTEGER,
	bom_format 			  String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
