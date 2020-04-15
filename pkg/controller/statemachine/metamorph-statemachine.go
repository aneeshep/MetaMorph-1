package controller

import (
	"bitbucket.com/metamorph/pkg/db/models/node"
	"fmt"
	"github.com/google/uuid"
	"runtime"
	"sync"
	"time"
)

type BMNode node.Node

type nodedb interface {
	GetNodes() ([]node.Node, error)
}

type DBHandler struct {
	db nodedb
}

func (db *DBHandler) GetNodes() ([]node.Node, error) {
	return node.GetNodes()
}

const (
	NEW          = "new"
	READY        = "ready"
	SETUPREADY   = "setupready"
	DEPLOYED     = "deployed"
	FAILED       = "failed"
	INTRANSITION = "in-transition"
)

type NodeStatus struct {
	NodeUUID uuid.UUID
	Status   bool
}

func (h *DBHandler) StartMetamorphFSM(runOnce bool) {

	fmt.Println("Starting Metamorph FSM")

	var wg sync.WaitGroup

	requestsChan := make(chan BMNode)
	nodeStatusChan := make(chan NodeStatus) //NodeUUID of nodes that did not make successful update to db
	//TODO : ensure sync.Oncce for these two go routines
	//TODO : Waitgroup for goroutines
	go checkFailedNodes(nodeStatusChan, &wg)
	go serviceRequest(requestsChan, nodeStatusChan, &wg)
	wg.Add(2)

	fmt.Println("Number of Go Routines ", runtime.NumGoroutine())

	for {

		//nodelist, err := node.GetNodes()
		nodelist, err := h.db.GetNodes()

		if err != nil {
			fmt.Println("Failed to get nodelist")

		} else {

			for _, bmnode := range nodelist {
				// What about nodes that are already in transistions.. should there be a transition state.
				fmt.Printf("[%v] - Starting Processing\n", bmnode.Name)
				requestsChan <- BMNode(bmnode)

			}
		}
		// set the array to nil for the next cycle
		nodelist = nil

		time.Sleep(10 * time.Millisecond) // sleep for 10 ms before the start of the next cycle
		if runOnce == true {              // for testing purpose only.
			break
		}
	}
	fmt.Println("Number of Go Routines", runtime.NumGoroutine())

	close(requestsChan)
	close(nodeStatusChan)
	wg.Wait()
	fmt.Println("Number of Go Routines", runtime.NumGoroutine())

}

func checkFailedNodes(nodeStatusChan chan NodeStatus, wg *sync.WaitGroup) {

	for nodestatus := range nodeStatusChan {

		if nodestatus.Status == false {
			fmt.Printf("Failed Node %v\n", nodestatus.NodeUUID)
			//try update of the db
		}
	}
	fmt.Println("Closing checkFailedNodes Goroutine")
	wg.Done()
}

func serviceRequest(requestsChan chan BMNode, nodeStatusChan chan<- NodeStatus, wg *sync.WaitGroup) {

	for bmnode := range requestsChan {

		switch bmnode.State {
		case NEW:
			fmt.Printf("[%v] Transition to Ready State\n", bmnode.Name)
			wg.Add(1)
			go ReadystateHandler(bmnode, nodeStatusChan, wg)
		case READY:
			fmt.Printf("[%v] Transitioning to SetupReady\n", bmnode.Name)
			wg.Add(1)
			go SetupreadyHandler(bmnode, nodeStatusChan, wg)
		case SETUPREADY:
			fmt.Printf("[%v] Transitioning to Deployed State\n", bmnode.Name)
			wg.Add(1)
			go DeployedHandler(bmnode, nodeStatusChan, wg)
		default:
			fmt.Printf("[%v] State not defined\n", bmnode.Name)
			//nodestatus := NodeStatus { NodeUUID: node.NodeUUID, Status: false }
			//nodeStatusChan <- nodestatus
		}

		fmt.Println("Number of Go Routines ", runtime.NumGoroutine())

	}
	fmt.Println("Closing serviceRequest Goroutine")
	wg.Done()

}

func ReadystateHandler(bmnode BMNode, nodeStatusChan chan<- NodeStatus, wg *sync.WaitGroup) {
	fmt.Printf("[%v] Entering Ready State Handler\n", bmnode.Name)
	bmnode.State = INTRANSITION
	fmt.Printf("[%v] - NodeUUID - %v\n", bmnode.Name, bmnode.NodeUUID)

	var testNode bool = true
	var nodestatus NodeStatus
	if testNode {
		fmt.Printf("[%v] Error condition\n", bmnode.Name)
		nodestatus = NodeStatus{NodeUUID: bmnode.NodeUUID, Status: false}
	}else {
		nodestatus = NodeStatus{NodeUUID: bmnode.NodeUUID, Status: true}
	}
	nodeStatusChan <- nodestatus

	wg.Done()

	//Do Ready Check verification
	//Update database accordingly

}

func SetupreadyHandler(bmnode BMNode, nodeStatusChan chan<- NodeStatus, wg *sync.WaitGroup) {
	fmt.Printf("[%v] Entering Setup Ready State Handler\n", bmnode.Name)
	bmnode.State = INTRANSITION
	fmt.Println(bmnode.NodeUUID)
	nodestatus := NodeStatus{NodeUUID: bmnode.NodeUUID, Status: true}
	nodeStatusChan <- nodestatus
	wg.Done()
	//Do Ready Check verification
	//Update database accordingly
}
func DeployedHandler(bmnode BMNode, nodeStatusChan chan<- NodeStatus, wg *sync.WaitGroup) {
	fmt.Printf("[%v] Entering Deployed State Handler\n", bmnode.Name)
	bmnode.State = INTRANSITION
	fmt.Println(bmnode.NodeUUID)
	nodestatus := NodeStatus{NodeUUID: bmnode.NodeUUID, Status: true}
	nodeStatusChan <- nodestatus
	wg.Done()
	//Do Ready Check verification
	//Update database accordingly
}