package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gobs/pretty"
	"github.com/zsuzhengdu/grafana-annotations/config"
	"github.com/zsuzhengdu/grafana-annotations/grafana"

	cmap "github.com/orcaman/concurrent-map"

	log "github.com/sirupsen/logrus"

	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/facebookgo/flagenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChartStatus represents status of a helm chart
type ChartStatus struct {
	NAME   string
	STATUS string
}

func remove(s []ChartStatus, index int) []ChartStatus {
	return append(s[:index], s[index+1:]...)
}

var (
	settings   = cli.New()
	clients    = cmap.New()
	namespaces = flag.String("namespaces", "", "namespaces to monitor.  Defaults to all")
	configFile = flag.String("config", "", "Configfile to load for helm overwrite registries.  Default is empty")

	grafanaIP   = flag.String("grafana-ip", "", "grafanaIP to send annotation. Default to 127.0.0.1")
	auth        = flag.String("auth", "", "grafana auth: can be in user:pass format, or it can be an api key. Default to admin:admin")
	dashboardID = flag.Int64("dashboard-id", 0, "dashboardID to send annotation. Default to 0")
	panelID     = flag.Int64("panel-id", 0, "panelD to send annotation. Default to 0")

	intervalDuration = flag.String("interval-duration", "10s", "Enable metrics gathering in background, each given duration. If not provided, the helm stats are computed synchronously.  Default is 10s")

	infoMetric      = flag.Bool("info-metric", true, "Generate info metric.  Defaults to true")
	timestampMetric = flag.Bool("timestamp-metric", true, "Generate timestamps metric.  Defaults to true")

	fetchLatest = flag.Bool("latest-chart-version", true, "Attempt to fetch the latest chart version from registries. Defaults to true")

	prometheusHandler = promhttp.Handler()

	// ChartsStatus stores the status of charts in namespaces
	ChartsStatus = make(map[string][]ChartStatus)
)

func initFlags() config.AppConfig {
	cliFlags := new(config.AppConfig)
	cliFlags.ConfigFile = *configFile
	return *cliFlags
}

func runStats(config config.Config) {
	// ns deleted and notify all deployment gone
	var keys []string
	for key := range ChartsStatus {
		keys = append(keys, key)
	}

	var nsList []string

	for ns := range clients.Items() {
		nsList = append(nsList, ns)
	}

	m := make(map[string]bool)
	for _, item := range nsList {
		m[item] = true
	}

	for _, key := range keys {
		if _, ok := m[key]; !ok {
			// annotate on all deployments in namespace: key
			delete(ChartsStatus, key)
		}
	}

	for ns, client := range clients.Items() {
		list := action.NewList(client.(*action.Configuration))
		items, err := list.Run()
		if err != nil {
			log.Warnf("got error while listing %v", err)
			continue
		}

		_, ok := ChartsStatus[ns]

		if !ok {
			for _, item := range items {
				ChartsStatus[ns] = append(ChartsStatus[ns], ChartStatus{item.Chart.Name(), item.Info.Status.String()})
				if !checkAnnotationAlreadyAdded(*dashboardID, *panelID, item.Info.FirstDeployed.UnixNano()/1000000, []string{item.Chart.Name(), ns}, item.Info.Status.String()) {
					log.Infof("Send annotation to deployed: %s in namespace %s", item.Chart.Name(), ns)
					addAnnotations(*dashboardID, *panelID, item.Info.FirstDeployed.UnixNano()/1000000, []string{item.Chart.Name(), ns}, item.Info.Status.String())
				}
			}
		} else {
			// removed undeployed charts in namespace
			var list []string
			for _, status := range ChartsStatus[ns] {
				list = append(list, status.NAME)
			}
			var listItems []string
			for _, item := range items {
				listItems = append(listItems, item.Chart.Name())
			}
			for _, l := range list {
				if !stringInSlice(l, listItems) {
					// remove undeployed chart from ChartsStatus in namespace ns
					var index int
					for i, val := range ChartsStatus[ns] {
						if val.NAME == l {
							index = i
							break
						}
					}
					log.Infof("Send annotation to undeployed: %s in namespace %s", ChartsStatus[ns][index].NAME, ns)
					addAnnotations(*dashboardID, *panelID, time.Now().UnixNano()/1000000, []string{ChartsStatus[ns][index].NAME, ns}, "Undeployed")
					ChartsStatus[ns] = remove(ChartsStatus[ns], index)
				}
			}
			// Delete namespace from ChartsStatus due to no deployments
			if len(ChartsStatus[ns]) == 0 {
				delete(ChartsStatus, ns)
			}

			for _, item := range items {
				if !stringInSlice(item.Chart.Name(), list) {
					ChartsStatus[ns] = append(ChartsStatus[ns], ChartStatus{item.Chart.Name(), item.Info.Status.String()})
					// send annotation for new deployment in the ns
					log.Infof("Send annotation to new deployed: %s in namespace %s", item.Chart.Name(), ns)
					addAnnotations(*dashboardID, *panelID, item.Info.FirstDeployed.UnixNano()/1000000, []string{item.Chart.Name(), ns}, item.Info.Status.String())
				} else {
					for index, status := range ChartsStatus[ns] {
						if item.Chart.Name() == status.NAME && item.Info.Status.String() != status.STATUS {
							ChartsStatus[ns][index].STATUS = item.Info.Status.String()
							// send annoation for deployment with status update
							log.Infof("Send annotation to %s in namespace %s with status update to %s", ChartsStatus[ns][index].NAME, ns, ChartsStatus[ns][index].STATUS)
							addAnnotations(*dashboardID, *panelID, item.Info.LastDeployed.UnixNano()/1000000, []string{ChartsStatus[ns][index].NAME, ns}, ChartsStatus[ns][index].STATUS)
						}
					}
				}
			}
		}
	}
}

func runStatsPeriodically(interval time.Duration, config config.Config) {
	for {
		runStats(config)
		time.Sleep(interval)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {

}

func connect(namespace string) {
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Infof)
	if err != nil {
		log.Warnf("failed to connect to %s with %v", namespace, err)
	} else {
		log.Infof("Watching namespace %s", namespace)
		clients.Set(namespace, actionConfig)
	}
}

func informer() {
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Infof)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := actionConfig.KubernetesClientSet()
	if err != nil {
		log.Fatal(err)
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Namespaces().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// "k8s.io/apimachinery/pkg/apis/meta/v1" provides an Object
			// interface that allows us to get metadata easily
			mObj := obj.(v1.Object)
			connect(mObj.GetName())
		},
		DeleteFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Infof("Removing namespace %s", mObj.GetName())
			clients.Remove(mObj.GetName())
		},
	})

	informer.Run(stopper)
}

func main() {
	flagenv.Parse()
	flag.Parse()
	cliFlags := initFlags()
	config := config.LoadConfiguration(cliFlags.ConfigFile)

	log.Infof("DashboardID: %d, PanelID: %d, intervalDuration: %s", *dashboardID, *panelID, *intervalDuration)

	runIntervalDuration, err := time.ParseDuration(*intervalDuration)
	if err != nil {
		log.Fatalf("invalid duration `%s`: %s", *intervalDuration, err)
	}

	if namespaces == nil || *namespaces == "" {
		go informer()
	} else {
		for _, namespace := range strings.Split(*namespaces, ",") {
			connect(namespace)
		}
	}

	if grafanaIP == nil || *grafanaIP == "" {
		getGrafanaIP()
	}

	if auth == nil || *auth == "" {
		getGrafanaAuth()
	}

	go runStatsPeriodically(runIntervalDuration, config)

	http.HandleFunc("/healthz", healthz)
	log.Fatal(http.ListenAndServe(":9571", nil))
}

func getGrafanaIP() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	// var grafanaIP string
	for _, item := range services.Items {
		if len(item.ObjectMeta.Labels) > 0 {
			if val, ok := item.ObjectMeta.Labels["app.kubernetes.io/name"]; ok && val == "grafana" {
				*grafanaIP = item.Spec.ClusterIP
			}
		}
	}
	log.Infof("In-Cluster GrafanaIP: %s", *grafanaIP)
	*grafanaIP = "http://" + *grafanaIP
}

func getGrafanaAuth() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secrets, err := clientset.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{LabelSelector: })
	if err != nil {
		panic(err.Error())
	}

	for _, item := range secrets.Items {
		if len(item.ObjectMeta.Labels) > 0 {
			if val, ok := item.ObjectMeta.Labels["app.kubernetes.io/name"]; ok && val == "grafana" {
				*auth = string(item.Data["admin-user"]) + ":" + string(item.Data["admin-password"])
			}
		}
	}
	log.Infof("auth: %s", *auth)
}

func checkAnnotationAlreadyAdded(DashboardID, PanelID, Time int64, Tags []string, Text string) bool {
	c, err := grafana.New(*auth, *grafanaIP)
	if err != nil {
		fmt.Printf("expected error to be nil; got: %s", err.Error())
	}

	params := url.Values{}
	params.Add("dashboardId", strconv.FormatInt(DashboardID, 10))
	params.Add("panelId", strconv.FormatInt(PanelID, 10))
	params.Add("from", strconv.FormatInt(Time, 10))
	params.Add("to", strconv.FormatInt(Time, 10))
	params.Add("tags", Tags[0])
	as, err := c.Annotations(params)
	if err != nil {
		log.Debugf("expected error to be nil; got: %s", err.Error())
	}

	for _, item := range as {
		if item.Text == Text {
			return true
		}
	}
	return false
}

func addAnnotations(DashboardID, PanelID, Time int64, Tags []string, Text string) {
	c, err := grafana.New(*auth, *grafanaIP)
	if err != nil {
		fmt.Printf("expected error to be nil; got: %s", err.Error())
	}

	a := grafana.Annotation{
		DashboardID: DashboardID,
		PanelID:     PanelID,
		Time:        Time,
		IsRegion:    true,
		TimeEnd:     Time,
		Tags:        Tags,
		Text:        Text,
	}

	// Better error handling when failing in adding annotations
	res, err := c.NewAnnotation(&a)
	if err != nil {
		fmt.Printf("expected error to be nil; got: %s", err.Error())
	}
	fmt.Println(pretty.PrettyFormat(res))
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
