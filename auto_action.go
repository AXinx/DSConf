package main

/**
 * Action smart contract.
 **/
import (
	"fmt"
	"encoding/json"
	"strconv"

	// April 2020, Updated for Fabric 2.0
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"

)

// Channel Name
const    Channel = "airlinechannel"
// Chaincode to be invoked
const    TargetChaincode = "token"

// TokenChaincode Represents our chaincode object
type Operators struct {
	OperatorID string 
	OclToken float64
}

type ActionProposal struct {
	Id 						int
	OrganisationId_proposal string
	OrganisationId_action   string 
	Action           		string
	Description				string
	Manual					int
	Done					int 
}

type ListAction struct {
	Id 						int
	OrganisationId_proposal string
	Action           		string
	Description				string
	Done					int 
}

// Init Implements the Init method
func (organisations *Operators) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Simply print a message
	fmt.Println("Initialize voting contract")

	// Init the datastructures
	actionProposal, _ := json.Marshal([]ActionProposal{})
	listAction, _ := json.Marshal([]ListAction{})

	stub.PutState("ActionProposal", actionProposal)
	stub.PutState("ListAction", listAction)

	// Return success
	return shim.Success([]byte("true"))
}

// Invoke method
func (organisations *Operators) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Get params
	funcName, args := stub.GetFunctionAndParameters()

	if(funcName == "actionProposal") {
		manual, _ := strconv.Atoi(args[3])
		return actionProposal(stub, args[0], args[1], args[2], manual)
	} else if(funcName == "listAction") {
		return listAction(stub, args[0], args[1])
	} else if(funcName == "executeAction") {

		if(len(args) == 1) {
			return executeAction(stub, 0, 0, args[0])
		}

		index, _ := strconv.Atoi(args[0])
		reply, _ := strconv.Atoi(args[1])

		return executeAction(stub, index, reply, "")
	} else if(funcName == "replyProposal") {
		fmt.Println("testjeee")
		index, _ := strconv.Atoi(args[0])
		reply, _ := strconv.Atoi(args[1])
		return replyAction(stub, index, reply)
	} else if(funcName == "get") {
		return get(stub, args[0])
	} else if(funcName == "reset") {
		return reset(stub)
	}

	return shim.Success([]byte("true"))
}

func actionProposal(stub shim.ChaincodeStubInterface, OrgId_action string, 
					Action string, Description string, Manual int) peer.Response {
	MSPid, _ := shim.GetMSPID()
	var newId int

	// Init vote receive list
	proposalList := []ActionProposal{}
	
	// Get all votes
	proposals, _ := stub.GetState("ActionProposal")
	json.Unmarshal(proposals, &proposalList)

	fmt.Println("Votes: " + string(len(proposalList)))
	if(len(proposalList) == 0) {
		newId = 0	
	} else {
		newId = proposalList[len(proposalList) - 1].Id + 1
	}

	new_proposal := ActionProposal{Id: newId, OrganisationId_proposal: MSPid, 
								   OrganisationId_action: OrgId_action, 
								   Action: Action, Description: Description,
								   Manual: Manual}

	proposalJson, _ := json.Marshal(append(proposalList, new_proposal))
	stub.PutState("ActionProposal", proposalJson)
	
	fmt.Println("Votes: " + string(proposals))
	return shim.Success([]byte("Action proposed to organisation: " + OrgId_action))
}

func listAction(stub shim.ChaincodeStubInterface, Action string, 
				Description string) peer.Response {
	MSPid, _ := shim.GetMSPID()
	var newId int

	// Init vote receive list
	proposalList := []ListAction{}
	
	// Get all votes
	proposals, _ := stub.GetState("ListAction")
	json.Unmarshal(proposals, &proposalList)

	if(len(proposalList) == 0) {
		newId = 0	
	} else {
		newId = proposalList[len(proposalList) - 1].Id + 1
	}

	new_proposal := ListAction{Id: newId, OrganisationId_proposal: MSPid, 
								   Action: Action, Description: Action}

	proposalJson, _ := json.Marshal(append(proposalList, new_proposal))
	stub.PutState("ListAction", proposalJson)
	
	return shim.Success([]byte("Action listed."))
}

func replyAction(stub shim.ChaincodeStubInterface, index int, 
				 reply int) peer.Response {
	MSPid, _ := shim.GetMSPID()

	// Init vote receive list
	actionProposal := []ListAction{}

	// Get all ActionProposal
	proposals, _ := stub.GetState("ListAction")
	json.Unmarshal(proposals, &actionProposal)
	
	// Check if somebody already replied on this proposal
	if(actionProposal[index].Done == 1) {
		return shim.Success([]byte("Already replied on this proposal"))
	}

	// Check if the operators has the right to reply
	if(actionProposal[index].OrganisationId_proposal == MSPid) {
		return_message := "Proposal has to be approved by other organisation."
		return shim.Success([]byte(return_message))
	}

	// Check reply
	if(reply == 1) {
		actionProposal[index].Done = 1

		colabProtocol(stub,
						actionProposal[index].OrganisationId_proposal, 
						MSPid,
						"proposal", "agree", 0)
	} else if(reply == 0) {
		actionProposal[index].Done = 1

		colabProtocol(stub,
						actionProposal[index].OrganisationId_proposal, 
						MSPid,
						"proposal", "disagree", 0)
	}


	actionJson, _ := json.Marshal(actionProposal)
	stub.PutState("ListAction", actionJson)

	var temp string
	if(reply == 0) {
		temp = "Disgree"
	} else {
		temp = "Agree"
	}

	return shim.Success([]byte("Replied on proposal: " + temp))
}

func executeAction(stub shim.ChaincodeStubInterface, index int, 
				   reply int, self_proposed string) peer.Response {
	MSPid, _ := shim.GetMSPID()

	// Init vote receive list
	actionProposal := []ActionProposal{}
	actionList := []ListAction{}
	
	// Get all votes
	listAction, _ := stub.GetState("ListAction")
	json.Unmarshal(listAction, &actionList)

	// // Get all ActionProposal
	proposals, _ := stub.GetState("ActionProposal")
	json.Unmarshal(proposals, &actionProposal)

	// Check if the execution is self proposed
	if(self_proposed != "") {
		// Self proposed 
		var newId int
		fmt.Println("VM scaled")

		if(len(actionProposal) == 0) {
			newId = 0	
		} else {
			newId = actionProposal[len(actionProposal) - 1].Id + 1
		}

		new_proposal := ActionProposal{Id: newId, 
									   OrganisationId_proposal: MSPid, 
									   OrganisationId_action: MSPid, 
									   Action: self_proposed, 
									   Description: "",
									   Manual: 0}
		
		actionProposal = append(actionProposal, new_proposal)	

		colabProtocol(stub, MSPid, "","proposal", "agree", 0)
	} else {
		// Check if there is already replied to this request
		if(actionProposal[index].Done == 1) {
			return shim.Success([]byte("Already replied to this request."))
		}
		
		// Not self proposed
		// Check if organisation is allowed to execute action
		if(actionProposal[index].OrganisationId_action != MSPid) {
			return shim.Success([]byte("Not allowed to take this actions."))
		}

		if(reply == 0) {
			colabProtocol(stub,
				actionProposal[index].OrganisationId_action, 
				actionProposal[index].OrganisationId_proposal,
				"proposal", "disagree", 0)

			return shim.Success([]byte("true")) 
		}

		// If proposal needs manual execution
		if(actionProposal[index].Manual == 1) {
			var newId int
			actionProposal[index].Done = 1

			colabProtocol(stub,
						actionProposal[index].OrganisationId_action, 
						actionProposal[index].OrganisationId_proposal,
						"proposal", "agree", 1)

			if(len(actionList) == 0) {
				newId = 0	
			} else {
				newId = actionList[len(actionList) - 1].Id + 1
			}

			listNewAction := ListAction{newId, actionProposal[index].OrganisationId_action,
										actionProposal[index].Action, 
										actionProposal[index].Description, 0}

			actionList = append(actionList, listNewAction)

			actionJson, _ := json.Marshal(actionProposal)
			stub.PutState("ActionProposal", actionJson)
		
			actionListJson, _ := json.Marshal(actionList)
			stub.PutState("ListAction", actionListJson)
		
			// More STD actions...
			return shim.Success([]byte("VM Scaled"))
		}
	}

	// Automatic action options
	if(actionProposal[index].Action == "scalevm") {
		// Make an API call to IaC service
		fmt.Println("VM scaled")

		if(self_proposed == "") {
			// Proposal done
			actionProposal[index].Done = 1

			colabProtocol(stub,
						actionProposal[index].OrganisationId_action, 
						actionProposal[index].OrganisationId_proposal,
						"proposal", "agree", 0)
		}
					  
	}

	actionJson, _ := json.Marshal(actionProposal)
	stub.PutState("ActionProposal", actionJson)

	actionListJson, _ := json.Marshal(actionList)
	stub.PutState("ListAction", actionListJson)

	// More STD actions...
	return shim.Success([]byte("VM Scaled"))
}

func get(stub shim.ChaincodeStubInterface, actionType string) peer.Response {
	// Holds a string for the response
	var actions string

	// Local variables for value & error
	var get_value  []byte
	var err    error
	
	if get_value, err = stub.GetState(actionType); err != nil {

		fmt.Println("Get Failed!!! ", err.Error())

		return shim.Error(("Get Failed!! "+err.Error()+"!!!"))

	} 

	// nil indicates non existent key
	if (get_value == nil) {
		actions = "-1"
	} else {
		actions = string(get_value)
	}

	fmt.Println(actions)
	
	return shim.Success([]byte(actions))
}

func colabProtocol(stub shim.ChaincodeStubInterface, org1 string, org2 string,
				   ac1 string, ac2 string, manual int) peer.Response {
	var payOffOrg1 float64
	var payOffOrg2 float64
	
	// Init Operator list
	Operator_list := []Operators{}

	// Get Operator list
	args := make([][]byte, 2)
	args[0] = []byte("get")
	args[1] = []byte("Operators")

	value := stub.InvokeChaincode(TargetChaincode, args, Channel)
	json.Unmarshal(value.Payload, &Operator_list)

	fmt.Println(Operator_list)

	// Check colab cases
	if(ac1 == "proposal" && ac2 == "agree") {
		payOffOrg1 = 2
		payOffOrg2 = 1
	} else if(ac1 == "proposal" && ac2 == "disagree") {
		payOffOrg1 = -1
		payOffOrg2 = 0.5
	} else if(ac1 == "performed" && ac2 == "agree") {
		payOffOrg1 = 2
		payOffOrg2 = 1
	} else if(ac1 == "performed" && ac2 == "disagree") {
		payOffOrg1 = 0
		payOffOrg2 = 0.5
	}

	// Edge cases
	if(manual == 1) {
		payOffOrg1 = 0
	}

	if(org2 == "") {
		payOffOrg2 = 0

		for i := 0; i < len(Operator_list); i++ {
			if(Operator_list[i].OperatorID == org1) {
				Operator_list[i].OclToken += payOffOrg1
			}
		}

	} else { 

		for i := 0; i < len(Operator_list); i++ {
			if(Operator_list[i].OperatorID == org1) {
				Operator_list[i].OclToken += payOffOrg1
			}

			if(Operator_list[i].OperatorID == org2) {
				Operator_list[i].OclToken += payOffOrg2
			}
		}

	}

	OperatorsJson, _ := json.Marshal(Operator_list)
	args1 := make([][]byte, 2)
	args1[0] = []byte("updateOCLtoken")
	args1[1] = []byte(OperatorsJson)

	stub.InvokeChaincode(TargetChaincode, args1, Channel)

	stub.PutState("Operators", OperatorsJson)
	return shim.Success([]byte("Test"))
}

func reset(stub shim.ChaincodeStubInterface) peer.Response {
	// Init the datastructures
	actionProposal, _ := json.Marshal([]ActionProposal{})
	listAction, _ := json.Marshal([]ListAction{})

	stub.PutState("ActionProposal", actionProposal)
	stub.PutState("ListAction", listAction)

	// Return success
	return shim.Success([]byte("true"))
}


// Chaincode registers with the Shim on startup
func main() {
	fmt.Printf("Started Chaincode.\n")
	err := shim.Start(new(Operators))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}