CREATE TABLE IF NOT EXISTS trivysbom (
	id                    UUID,
	image_name            String,
	package_url           String,
	bom_ref 			  String,
	serial_number         String,
	version 			  INTEGER
	bom_format 			  String,
	component_version     String,
	component_mime_type   String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
