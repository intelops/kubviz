CREATE TABLE IF NOT EXISTS kubescore (
	id 				UUID,
	clustername 	String,
	object_name 	String,
	kind 			String,
	apiVersion 		String,
	name 			String,
	namespace 		String,
	target_type 	String,
	description 	String,
	path 			String,
	summary 		String,
	file_name 		String,
	file_row  		String,
    EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

