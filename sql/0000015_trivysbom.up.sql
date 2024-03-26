CREATE TABLE IF NOT EXISTS trivysbom (
	id                    UUID,
	cluster_name 		  String,
	bom_format 			  String,
	serial_number         String,
	bom_ref 			  String,
	image_name            String,
	componet_type 		  String,
	package_url           String,
	time_stamp 			  DateTime('UTC'),
	other_component_name  String,
	other_component_bomref String,
	other_component_type String,
	other_component_version String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}},
	ExportedAt DateTime DEFAULT NULL
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
