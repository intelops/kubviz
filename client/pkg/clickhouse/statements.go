package clickhouse

type DBStatement string

const kubvizTable DBStatement = `
	CREATE TABLE IF NOT EXISTS events (
		id           UUID,
		op_type      String,
		name         String,
		namespace    String,
		kind         String,
		message      String,
		reason       String,
		host         String,
		event        String,
		first_time   DateTime,
		last_time    DateTime,
		event_time   DateTime,
		cluster_name String
	) engine=File(TabSeparated)
`
const rakeesTable DBStatement = `
CREATE TABLE IF NOT EXISTS rakkess (
	ClusterName String,
	Name String,
	Create String,
	Delete String,
	List String,
	Update String
) engine=File(TabSeparated)
`
const kubePugDepricatedTable DBStatement = `
CREATE TABLE IF NOT EXISTS DeprecatedAPIs (
	ClusterName String,
	Description String,
	Kind String,
	Deprecated UInt8,
	Scope String,
	ObjectName String
) engine=File(TabSeparated)
`
const kubepugDeletedTable DBStatement = `
CREATE TABLE IF NOT EXISTS DeletedAPIs (
	ClusterName String,
	Group String,
	Kind String,
	Version String,
	Name String,
	Deleted UInt8,
	Scope String,
	ObjectName String
) engine=File(TabSeparated)
`
const ketallTable DBStatement = `
CREATE TABLE IF NOT EXISTS getall_resources (
	Cluster_Name String,
	Namespace String,
	Kind String,
	Resource String,
	Age String
) engine=File(TabSeparated)
`
const outdateTable DBStatement = `
CREATE TABLE IF NOT EXISTS outdated_images (
	Cluster_Name String,
	Namespace String,
	Pod String,
	Current_Image String,
	Current_Tag String,
	Latest_Version String,
	Versions_Behind Int64
) engine=File(TabSeparated)
`
const InsertRakees DBStatement = "INSERT INTO rakkess (ClusterName, Name, Create, Delete, List, Update) VALUES (?, ?, ?, ?, ?, ?)"
const InsertKetall DBStatement = "INSERT INTO getall_resources (Cluster_Name, Namespace, Kind, Resource, Age) VALUES (?, ?, ?, ?, ?)"
const InsertOutdated DBStatement = "INSERT INTO outdated_images (Cluster_Name, Namespace, Pod, Current_Image, Current_Tag, Latest_Version, Versions_Behind) VALUES (?, ?, ?, ?, ?, ?, ?)"
const InsertDepricatedApi DBStatement = "INSERT INTO DeprecatedAPIs (ClusterName, Description, Kind, Deprecated, Scope, ObjectName) VALUES (?, ?, ?, ?, ?, ?)"
const InsertDeletedApi DBStatement = "INSERT INTO DeletedAPIs (ClusterName, Group, Kind, Version, Name, Deleted, Scope, ObjectName) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
const InsertKubvizEvent DBStatement = "INSERT INTO events (id, op_type, name, namespace, kind, message, reason, host, event, first_time, last_time, event_time, cluster_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
const clickhouseExperimental DBStatement = `SET allow_experimental_object_type=1;`
const containerTable DBStatement = `CREATE table IF NOT EXISTS container_bridge(event JSON) ENGINE = MergeTree ORDER BY tuple();`
const gitTable DBStatement = `CREATE table IF NOT EXISTS git_json(event JSON) ENGINE = MergeTree ORDER BY tuple();`
