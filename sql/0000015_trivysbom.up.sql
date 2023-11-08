CREATE TABLE IF NOT EXISTS trivysbom (
	id                    UUID,
	image_name            String,
	image_version         String,
	package_url           String,
	mime_type             String,
	bom_ref 			  String,
	serial_number         String,
	version 			  INTEGER
	bom_format 			  String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
