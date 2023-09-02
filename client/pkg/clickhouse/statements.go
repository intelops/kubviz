package clickhouse

type DBStatement string

const kubvizTable DBStatement = `
	CREATE TABLE IF NOT EXISTS events (
		ClusterName String,
		Id          String,
		EventTime   DateTime('UTC'),
		OpType      String,
		Name         String,
		Namespace    String,
		Kind         String,
		Message      String,
		Reason       String,
		Host         String,
		Event        String,
		FirstTime   String,
		LastTime    String
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
	ObjectName String,
	Description String,
	Kind String,
	Deprecated UInt8,
	Scope String
) engine=File(TabSeparated)
`
const kubepugDeletedTable DBStatement = `
CREATE TABLE IF NOT EXISTS DeletedAPIs (
	ClusterName String,
	ObjectName String,
	Group String,
	Kind String,
	Version String,
	Name String,
	Deleted UInt8,
	Scope String
) engine=File(TabSeparated)
`
const ketallTable DBStatement = `
CREATE TABLE IF NOT EXISTS getall_resources (
	ClusterName String,
	Namespace String,
	Kind String,
	Resource String,
	Age String
) engine=File(TabSeparated)
`
const outdateTable DBStatement = `
CREATE TABLE IF NOT EXISTS outdated_images (
	ClusterName String,
	Namespace String,
	Pod String,
	CurrentImage String,
	CurrentTag String,
	LatestVersion String,
	VersionsBehind Int64
) engine=File(TabSeparated)
`
const kubescoreTable DBStatement = `
	    CREATE TABLE IF NOT EXISTS kubescore (
		    id UUID,
			namespace String,
			cluster_name String,
			recommendations String
	    ) engine=File(TabSeparated)
	`
const trivyTableVul DBStatement = `
	    CREATE TABLE IF NOT EXISTS trivy_vul (
		    id UUID,
			cluster_name String,
			namespace String,
			kind String,
			name String,
            vul_id String,
            vul_vendor_ids String,
			vul_pkg_id String,
			vul_pkg_name String,
			vul_pkg_path String,
            vul_installed_version String,
            vul_fixed_version String,
			vul_title String,
			vul_severity String,
			vul_published_date DateTime('UTC'),
			vul_last_modified_date DateTime('UTC')
	    ) engine=File(TabSeparated)
	`

const trivyTableMisconfig DBStatement = `
	    CREATE TABLE IF NOT EXISTS trivy_misconfig (
		    id UUID,
			cluster_name String,
			namespace String,
			kind String,
			name String,
			misconfig_id String,
			misconfig_avdid String,
			misconfig_type String,
            misconfig_title String,
            misconfig_desc String,
			misconfig_msg String,
			misconfig_query String,
			misconfig_resolution String,
			misconfig_severity String,
			misconfig_status String
	    ) engine=File(TabSeparated)
	`
const trivyTableImage DBStatement = `
	CREATE TABLE IF NOT EXISTS trivyimage (
		id UUID,
		cluster_name String,
		artifact_name String,
		vul_id String,
		vul_pkg_id String,
		vul_pkg_name String,
		vul_installed_version String,
		vul_fixed_version String,
		vul_title String,
		vul_severity String,
		vul_published_date DateTime('UTC'),
		vul_last_modified_date DateTime('UTC')
	) engine=File(TabSeparated)
	`
const dockerHubBuildTable DBStatement = `
	CREATE TABLE IF NOT EXISTS dockerhubbuild (
		PushedBy String,
		ImageTag String,
		RepositoryName String,
		DateCreated String,
		Owner String,
		Event String
	) engine=File(TabSeparated)
	`
const azureContainerPushEventTable DBStatement = `
	CREATE TABLE IF NOT EXISTS azurecontainerpush (
		RegistryURL String,
		RepositoryName String,
		Tag String,
		ImageName String,
		Event String,
		Timestamp String,
		Size Int32,
		SHAID String
	) engine=File(TabSeparated)
	`
const quayContainerPushEventTable DBStatement = `
	CREATE TABLE IF NOT EXISTS quaycontainerpush (
		dockerURL String,
		repository String,
		tag String,
		name String,
		Event String,
		nameSpace String,
		homePage String
	) engine=File(TabSeparated)
	`
const trivySbomTable DBStatement = `
	CREATE TABLE IF NOT EXISTS trivysbom (
		id UUID,
		schema String,
		bom_format String,
		spec_version String,
		serial_number String,
		version INTEGER,
		metadata_timestamp DateTime('UTC'),
		metatool_vendor String,
		metatool_name String,
		metatool_version String,
		component_bom_ref String,
		component_type String,
		component_name String,
		component_version String,
		component_property_name String,
		component_property_value String,
		component_hash_alg String,
		component_hash_content String,
		component_license_exp String,
		component_purl String,
		dependency_ref String
	) engine=File(TabSeparated)
	`

const InsertDockerHubBuild DBStatement = "INSERT INTO dockerhubbuild (PushedBy, ImageTag, RepositoryName, DateCreated, Owner, Event) VALUES (?, ?, ?, ?, ?, ?)"
const InsertRakees DBStatement = "INSERT INTO rakkess (ClusterName, Name, Create, Delete, List, Update) VALUES (?, ?, ?, ?, ?, ?)"
const InsertKetall DBStatement = "INSERT INTO getall_resources (ClusterName, Namespace, Kind, Resource, Age) VALUES (?, ?, ?, ?, ?)"
const InsertOutdated DBStatement = "INSERT INTO outdated_images (ClusterName, Namespace, Pod, CurrentImage, CurrentTag, LatestVersion, VersionsBehind) VALUES (?, ?, ?, ?, ?, ?, ?)"
const InsertDepricatedApi DBStatement = "INSERT INTO DeprecatedAPIs (ClusterName, ObjectName, Description, Kind, Deprecated, Scope) VALUES (?, ?, ?, ?, ?, ?)"
const InsertDeletedApi DBStatement = "INSERT INTO DeletedAPIs (ClusterName, ObjectName, Group, Kind, Version, Name, Deleted, Scope) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
const InsertKubvizEvent DBStatement = "INSERT INTO events (ClusterName, Id, EventTime, OpType, Name, Namespace, Kind, Message, Reason, Host, Event, FirstTime, LastTime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
const clickhouseExperimental DBStatement = `SET allow_experimental_object_type=1;`
const containerDockerhubTable DBStatement = `CREATE table IF NOT EXISTS container_dockerhub(event JSON) ENGINE = MergeTree ORDER BY tuple();`
const containerGithubTable DBStatement = `CREATE table IF NOT EXISTS container_github(event JSON) ENGINE = MergeTree ORDER BY tuple();`
const InsertKubeScore string = "INSERT INTO kubescore (id, namespace, cluster_name, recommendations) VALUES (?, ?, ?, ?)"
const InsertTrivyVul string = "INSERT INTO trivy_vul (id, cluster_name, namespace, kind, name, vul_id, vul_vendor_ids, vul_pkg_id, vul_pkg_name, vul_pkg_path, vul_installed_version, vul_fixed_version, vul_title, vul_severity, vul_published_date, vul_last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?. ?)"
const InsertTrivyImage string = "INSERT INTO trivyimage (id, cluster_name, artifact_name, vul_id,  vul_pkg_id, vul_pkg_name,  vul_installed_version, vul_fixed_version, vul_title, vul_severity, vul_published_date, vul_last_modified_date) VALUES ( ?, ?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
const InsertTrivyMisconfig string = "INSERT INTO trivy_misconfig (id, cluster_name, namespace, kind, name, misconfig_id, misconfig_avdid, misconfig_type, misconfig_title, misconfig_desc, misconfig_msg, misconfig_query, misconfig_resolution, misconfig_severity, misconfig_status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?. ?, ?)"
const InsertAzureContainerPushEvent DBStatement = "INSERT INTO azurecontainerpush (RegistryURL, RepositoryName, Tag, ImageName, Event, Timestamp, Size, SHAID) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
const InsertTrivySbom string = "INSERT INTO trivysbom (id, schema, bom_format,spec_version,serial_number,  version, metadata_timestamp,metatool_vendor,metatool_name,metatool_version,component_bom_ref,component_type,component_name,component_version,component_property_name,component_property_value,component_hash_alg,component_hash_content,component_license_exp,component_purl,dependency_ref) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
const InsertQuayContainerPushEvent DBStatement = "INSERT INTO quaycontainerpush (dockerURL, repository, tag, name, Event, nameSpace, homePage) VALUES (?, ?, ?, ?, ?, ?, ?)"
