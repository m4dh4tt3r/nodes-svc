package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type nodeInfo struct {
	name     string
	podCount uint64
}

func getPodsPerNode(clientset *kubernetes.Clientset) (map[string]uint64, error) {
	podsPerNodeMap := make(map[string]uint64)

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, p := range pods.Items {
		podsPerNodeMap[p.Spec.NodeName]++
	}

	return podsPerNodeMap, nil
}

func sortNodesByPodCount(nodes map[string]uint64) []string {
	var nodesByPodCountSlice []nodeInfo
	var sortedNodes []string

	for k, v := range nodes {
		nodesByPodCountSlice = append(nodesByPodCountSlice, nodeInfo{k, v})
	}

	sort.Slice(nodesByPodCountSlice, func(i, j int) bool {
		return nodesByPodCountSlice[i].podCount > nodesByPodCountSlice[j].podCount
	})

	for _, n := range nodesByPodCountSlice {
		sortedNodes = append(sortedNodes, n.name)
	}

	return sortedNodes
}

func podHandler(w http.ResponseWriter, r *http.Request) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	podsPerNode, err := getPodsPerNode(clientset)
	if err != nil {
		panic(err.Error())
	}

	nodes := sortNodesByPodCount(podsPerNode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nodes)
}

func main() {
	http.HandleFunc("/", podHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}
