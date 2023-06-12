package rakkess

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/corneliusweig/tabwriter"
	"github.com/pkg/errors"
	"golang.org/x/term"
	authorizationv1 "k8s.io/api/authorization/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	v1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/klog/v2"
)

type ResourceAccess map[string]map[string]Access

type Access uint8

// This encodes the access of the given subject to the resource+verb combination.
const (
	Denied Access = iota
	Allowed
	NotApplicable
	RequestErr
)

const (
	FlagVerbs          = "verbs"
	FlagServiceAccount = "sa"
	FlagOutput         = "output"
	FlagVerbosity      = "verbosity"
	FlagDiffWith       = "diff-with"
)

var (
	// ValidVerbs is the list of allowed actions on kubernetes resources.
	// Sort order aligned along CRUD.
	ValidVerbs = []string{
		"create",
		"get",
		"list",
		"watch",
		"update",
		"patch",
		"delete",
		"deletecollection",
	}

	// ValidOutputFormats is the list of valid formats for the result table.
	ValidOutputFormats = []string{
		"icon-table",
		"ascii-table",
	}
)
var (
	// for testing
	getDiscoveryClient = getDiscoveryClientImpl
)

type GroupResource struct {
	APIGroup    string
	APIResource metav1.APIResource
}

func (ra ResourceAccess) Table(verbs []string) *Table {
	var names []string
	for name := range ra {
		names = append(names, name)
	}
	sort.Strings(names)

	// table header
	headers := []string{"NAME"}
	for _, v := range verbs {
		headers = append(headers, strings.ToUpper(v))
	}

	p := TableWithHeaders(headers)

	// table body
	for _, name := range names {
		var outcomes []Outcome

		res := ra[name]
		for _, v := range verbs {
			var o Outcome
			switch res[v] {
			case Denied:
				o = Down
			case Allowed:
				o = Up
			case NotApplicable:
				o = None
			case RequestErr:
				o = Err
			}
			outcomes = append(outcomes, o)
		}
		p.AddRow([]string{name}, outcomes...)
	}
	return p
}

// Extracts the full name including APIGroup, e.g. 'deployment.apps'
func (g GroupResource) fullName() string {
	if g.APIGroup == "" {
		return g.APIResource.Name
	}
	return fmt.Sprintf("%s.%s", g.APIResource.Name, g.APIGroup)
}

// FetchAvailableGroupResources fetches a list of known APIResources on the server.
func FetchAvailableGroupResources(opts *RakkessOptions) ([]GroupResource, error) {
	client, err := getDiscoveryClient(opts)
	if err != nil {
		return nil, errors.Wrap(err, "discovery client")
	}

	client.Invalidate()

	var resourcesFetcher func() ([]*metav1.APIResourceList, error)
	if opts.ConfigFlags.Namespace == nil || *opts.ConfigFlags.Namespace == "" {
		resourcesFetcher = client.ServerPreferredResources
	} else {
		resourcesFetcher = client.ServerPreferredNamespacedResources
	}

	resources, err := resourcesFetcher()
	if err != nil {
		if resources == nil {
			return nil, errors.Wrap(err, "get preferred resources")
		}
		klog.Warningf("Could not fetch full list of resources, result will be incomplete: %s", err)
	}

	var grs []GroupResource
	for _, list := range resources {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			klog.Warningf("Cannot parse groupVersion: %s", err)
			continue
		}
		for _, r := range list.APIResources {
			if len(r.Verbs) == 0 {
				continue
			}

			grs = append(grs, GroupResource{
				APIGroup:    gv.Group,
				APIResource: r,
			})
		}
	}

	return grs, nil
}

func getDiscoveryClientImpl(opts *RakkessOptions) (discovery.CachedDiscoveryInterface, error) {
	return opts.DiscoveryClient()
}
func CheckResourceAccess(ctx context.Context, sar authv1.SelfSubjectAccessReviewInterface, grs []GroupResource, verbs []string, namespace *string) ResourceAccess {
	var mu sync.Mutex // guards res
	res := make(ResourceAccess)

	var ns string
	if namespace != nil {
		ns = *namespace
	}

	var wg sync.WaitGroup
	for _, gr := range grs {
		wg.Add(1)
		// copy captured variables"github.com/corneliusweig/rakkess/pkg/rakkess/client"
		namespace := ns
		gr := gr
		go func() {
			defer wg.Done()

			klog.V(2).Infof("Checking access for %s", gr.fullName())

			// This seems to be a bug in kubernetes. If namespace is set for non-namespaced
			// resources, the access is reported as "allowed", but in fact it is forbidden.
			if !gr.APIResource.Namespaced {
				namespace = ""
			}

			allowedVerbs := sets.NewString(gr.APIResource.Verbs...)

			access := make(map[string]Access)
			for _, v := range verbs {
				if !allowedVerbs.Has(v) {
					access[v] = NotApplicable
					continue
				}

				req := authorizationv1.SelfSubjectAccessReview{
					Spec: authorizationv1.SelfSubjectAccessReviewSpec{
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Verb:      v,
							Resource:  gr.APIResource.Name,
							Group:     gr.APIGroup,
							Namespace: namespace,
						},
					},
				}

				var a Access
				resp, err := sar.Create(ctx, &req, metav1.CreateOptions{})
				switch {
				case err != nil:
					a = RequestErr
				case resp.Status.Allowed:
					a = Allowed
				}
				access[v] = a
			}

			mu.Lock()
			res[gr.fullName()] = access
			mu.Unlock()
		}()
	}

	wg.Wait()

	return res
}
func Diff(left, right ResourceAccess, verbs []string) *Table {
	// table header
	headers := []string{"NAME"}
	for _, v := range verbs {
		headers = append(headers, strings.ToUpper(v))
	}

	names := make([]string, 0, len(left))
	for name := range left {
		names = append(names, name)
	}
	sort.Strings(names)

	p := TableWithHeaders(headers)

	for _, name := range names {
		l, r := left[name], right[name]
		klog.V(3).Infof("left=%v right=%v name=%s", l, r, name)

		skip := true
		var outcomes []Outcome
		for _, verb := range verbs {
			ll, rr := l[verb], r[verb]
			var o Outcome
			if ll != rr {
				skip = false
				if ll == Allowed {
					o = Down
				}
				if rr == Allowed {
					o = Up
				}
			}
			outcomes = append(outcomes, o)
		}
		if !skip {
			p.AddRow([]string{name}, outcomes...)
		}
	}

	for name := range right {
		if _, ok := left[name]; !ok {
			klog.Warning("Some differences may be hidden, please swap the roles to get the full picture.")
			break
		}
	}

	return p
}

type RakkessOptions struct {
	ConfigFlags      *genericclioptions.ConfigFlags
	Verbs            []string
	AsServiceAccount string
	OutputFormat     string
	Streams          *genericclioptions.IOStreams
	ResourceList     []*metav1.APIResourceList
}

// NewRakkessOptions creates RakkessOptions with defaults.
func NewRakkessOptions() *RakkessOptions {
	return &RakkessOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(false),
		Streams: &genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

// Sets up options with in-memory buffers as in- and output-streams
func NewTestRakkessOptions() (*RakkessOptions, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	iostreams, in, out, errout := genericclioptions.NewTestIOStreams()
	klog.SetOutput(errout)
	return &RakkessOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		Streams:     &iostreams,
	}, in, out, errout
}

// GetAuthClient creates a client for SelfSubjectAccessReviews with high queries per second.
func (o *RakkessOptions) GetAuthClient() (v1.SelfSubjectAccessReviewInterface, error) {
	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	restConfig.QPS = 500
	restConfig.Burst = 1000

	authClient := v1.NewForConfigOrDie(restConfig)
	return authClient.SelfSubjectAccessReviews(), nil
}

// DiscoveryClient creates a kubernetes discovery client.
func (o *RakkessOptions) DiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return o.ConfigFlags.ToDiscoveryClient()
}

func (o *RakkessOptions) ExpandServiceAccount() error {
	if o.AsServiceAccount == "" {
		return nil
	}

	qualifiedServiceAccount, err := o.namespacedServiceAccount()
	if err != nil {
		return err
	}

	impersonate := fmt.Sprintf("system:serviceaccount:%s", qualifiedServiceAccount)
	klog.V(2).Infof("Impersonating as %s", impersonate)
	o.ConfigFlags.Impersonate = &impersonate
	return nil
}

func (o *RakkessOptions) namespacedServiceAccount() (string, error) {
	if strings.Contains(o.AsServiceAccount, ":") {
		return o.AsServiceAccount, nil
	}

	if o.ConfigFlags.Namespace != nil && *o.ConfigFlags.Namespace != "" {
		return fmt.Sprintf("%s:%s", *o.ConfigFlags.Namespace, o.AsServiceAccount), nil
	}

	return "", fmt.Errorf("serviceAccounts are namespaced, either provide --namespace or fully qualify the serviceAccount: '<namespace>:%s'", o.AsServiceAccount)
}

// ExpandVerbs expands wildcard verbs `*` and `all`.
func (o *RakkessOptions) ExpandVerbs() {
	for _, verb := range o.Verbs {
		if verb == "*" || verb == "all" {
			o.Verbs = ValidVerbs
		}
	}
}

type color int

const (
	red    = color(31)
	green  = color(32)
	purple = color(35)
	none   = color(0)
)

var (
	isTerminal = isTerminalImpl
	once       sync.Once
)

type Outcome uint8

const (
	None Outcome = iota
	Up
	Down
	Err
)

type Row struct {
	Intro   []string
	Entries []Outcome
}
type Table struct {
	Headers []string
	Rows    []Row
}

func initTerminal(_ io.Writer) {
}

func isTerminalImpl(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

func TableWithHeaders(headers []string) *Table {
	return &Table{
		Headers: headers,
	}
}

func (p *Table) AddRow(intro []string, outcomes ...Outcome) {
	row := Row{
		Intro:   intro,
		Entries: outcomes,
	}
	p.Rows = append(p.Rows, row)
}

func (p *Table) Render(out io.Writer, outputFormat string) {
	once.Do(func() { initTerminal(out) })

	conv := HumanreadableAccessCode
	if isTerminal(out) {
		conv = colored(conv)
	}
	if outputFormat == "ascii-table" {
		conv = asciiAccessCode
	}

	w := tabwriter.NewWriter(out, 4, 8, 2, ' ', tabwriter.SmashEscape|tabwriter.StripEscape)
	defer w.Flush()

	// table header
	for i, h := range p.Headers {
		if i == 0 {
			fmt.Fprint(w, h)
		} else {
			fmt.Fprintf(w, "\t%s", h)
		}
	}
	fmt.Fprint(w, "\n")

	// table body
	for _, row := range p.Rows {
		fmt.Fprintf(w, "%s", strings.Join(row.Intro, "\t"))
		for _, e := range row.Entries {
			fmt.Fprintf(w, "\t%s", conv(e)) // FIXME
		}
		fmt.Fprint(w, "\n")
	}
}

func HumanreadableAccessCode(o Outcome) string {
	switch o {
	case None:
		return ""
	case Up:
		return "✔" // ✓
	case Down:
		return "✖" // ✕
	case Err:
		return "ERR"
	default:
		panic("unknown access code")
	}
}

func colored(wrap func(Outcome) string) func(Outcome) string {
	return func(o Outcome) string {
		c := none
		switch o {
		case Up:
			c = green
		case Down:
			c = red
		case Err:
			c = purple
		}
		return fmt.Sprintf("\xff\033[%dm\xff%s\xff\033[0m\xff", c, wrap(o))
	}
}

func asciiAccessCode(o Outcome) string {
	switch o {
	case None:
		return "n/a"
	case Up:
		return "yes"
	case Down:
		return "no"
	case Err:
		return "ERR"
	default:
		panic("unknown access code")
	}
}

func Options(opts *RakkessOptions) error {
	if err := verbs(opts.Verbs); err != nil {
		return err
	}
	return OutputFormat(opts.OutputFormat)
}

func OutputFormat(format string) error {
	for _, o := range ValidOutputFormats {
		if o == format {
			return nil
		}
	}
	return fmt.Errorf("unexpected output format: %s", format)
}

func verbs(verbs []string) error {
	valid := sets.NewString(ValidVerbs...)
	given := sets.NewString(verbs...)
	difference := given.Difference(valid)

	if difference.Len() > 0 {
		return fmt.Errorf("unexpected verbs: %s", difference.List())
	}

	return nil
}

var version, gitCommit, buildDate string
var platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

// BuildInfo stores static build information about the binary.
type BuildInfo struct {
	BuildDate string
	Compiler  string
	GitCommit string
	GoVersion string
	Platform  string
	Version   string
}

// GetBuildInfo returns build information about the binary
func GetBuildInfo() *BuildInfo {
	// These vars are set via -ldflags settings during 'go build'
	return &BuildInfo{
		Version:   version,
		GitCommit: gitCommit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  platform,
	}
}

// ParseVersion parses a version string ignoring a leading `v`. For example: v1.2.3
func ParseVersion(version string) (semver.Version, error) {
	version = strings.TrimLeft(strings.TrimSpace(version), "v")
	return semver.Parse(version)
}
func Resource(ctx context.Context, opts *RakkessOptions) (ResourceAccess, error) {
	if err := Options(opts); err != nil {
		return nil, err
	}

	grs, err := FetchAvailableGroupResources(opts)
	if err != nil {
		return nil, errors.Wrap(err, "fetch available group resources")
	}
	klog.V(2).Info(grs)

	authClient, err := opts.GetAuthClient()
	if err != nil {
		return nil, errors.Wrap(err, "get auth client")
	}

	ret := CheckResourceAccess(ctx, authClient, grs, opts.Verbs, opts.ConfigFlags.Namespace)
	return ret, nil
}
