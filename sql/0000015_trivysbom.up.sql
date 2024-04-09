CREATE TABLE IF NOT EXISTS trivysbom (
	id                    UUID,
	cluster_name          String,
	bom_format 			  String,
	serial_number         String,
	bom_ref 			  String,
	image_name            String,
	component_type 		  String,
	package_url           String,
	event_time			  DateTime('UTC'),
	other_component_name  String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}},
	ExportedAt DateTime DEFAULT NULL
) ENGINE = MergeTree()
ORDER BY ExpiryDate
TTL ExpiryDate;
