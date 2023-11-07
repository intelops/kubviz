CREATE TABLE IF NOT EXISTS trivy_misconfig (
	id                 UUID,
	cluster_name       String,
	namespace          String,
	kind               String,
	name               String,
	misconfig_id       String,
	misconfig_avdid    String,
	misconfig_type     String,
	misconfig_title    String,
	misconfig_desc     String,
	misconfig_msg      String,
	misconfig_query    String,
	misconfig_resolution String,
	misconfig_severity String,
	misconfig_status   String,
    EventTime          DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

