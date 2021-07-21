package main

/**
 * Voting mechanism: Global proposals.
 **/
import (
	"fmt"
	"encoding/json"
	"time"
	"strconv"

	// April 2020, Updated for Fabric 2.0
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"

)

// TokenChaincode Represents our chaincode object
type Operators struct {
	OperatorID string 
	OclToken int
}

// 
type Votes struct {
	Id int
	CreatorID string
	Title string 
	Timestamp time.Time
	Duration string
	Description string
	Yes []Operators
	No []Operators
	Answer string
}

// Init Implements the Init method
func (Operator *Operators) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Simply print a message
	fmt.Println("Initialize voting contract")

	// Init lists
	voter_list, _ := json.Marshal([]Operators{})
	votes, _ := json.Marshal([]Votes{})
	closedVotes, _ := json.Marshal([]Votes{})

	// Put states to level DB
	stub.PutState("Operators", voter_list)
	stub.PutState("Votes", votes)
	stub.PutState("closedVotes", closedVotes)

	// Return success
	return shim.Success([]byte("true"))
}

// Invoke method
func (Operators *Operators) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	register(stub)

	// Get params
	funcName, args := stub.GetFunctionAndParameters()
	
	// Function invocation 
	if(funcName == "register" ){
		return register(stub)
	} else if (funcName == "createVote") {
		if( len(args) < 3) {
			return shim.Success([]byte("Need more args for createVote."))
		}

		return createVote(stub, args[0], args[1], args[2])
	} else if (funcName == "replyVote") {
		if ( len(args) < 2) {
			return shim.Success([]byte("Need more args for replyVote."))
		}

		return replyVote(stub, args[0], args[1])
	} else if(funcName == "get") {
		if( len(args) < 1) { 
			return shim.Success([]byte("Need more args for closeVote."))
		}

		return get(stub, args[0])
	} else if (funcName == "closeVotes") {
		if( len(args) < 1) { 
			return shim.Success([]byte("Need more args for closeVote."))
		}

		return closeVotes(stub, args[0])
	} else if (funcName == "updateOCLtoken") {
		stub.PutState("Operators", []byte(args[0]))
	} else if (funcName == "reset") {
		return reset(stub)
	}

	// Check lists
	value, _ := stub.GetState("Operators")
	fmt.Println("Voters: " + string(value))
	
	votes, _ := stub.GetState("Votes")
	fmt.Println("Votes: " + string(votes))

	return shim.Success([]byte("true"))
}

// Register to voting mechanism on current channel 
func register(stub shim.ChaincodeStubInterface) peer.Response {
		MSPid, _ := shim.GetMSPID()
		// Init datatypes
		voter_list := []Operators{}

		// Get Operator list
		value, _ := stub.GetState("Operators")
		json.Unmarshal(value, &voter_list)

		// Get current Operator
		for i := 0; i < len(voter_list); i++ {
			if(voter_list[i].OperatorID == MSPid) {
				return shim.Success([]byte("User already registerd"))
			}
		}

		Operator := Operators{OperatorID: MSPid, OclToken: 200}
		urlsJson, _ := json.Marshal(append(voter_list, Operator))

		stub.PutState("Operators", urlsJson)

		return shim.Success(urlsJson)
}

// Create a vote
func createVote(stub shim.ChaincodeStubInterface, title string, 
				duration string, description string) peer.Response {
		// Get organisation ID
		MSPid, _ := shim.GetMSPID()

		// Init vote receive list
		votes_list := []Votes{}
		
		// Get all votes
		votes, _ := stub.GetState("Votes")
		json.Unmarshal(votes, &votes_list)

		new_vote := Votes{Id: len(votes_list), CreatorID: MSPid, Title: title, 
						  Timestamp: time.Now(), Duration: duration, 
						  Description: description, Yes: []Operators{}, No: []Operators{}}

		
		votesJson, _ := json.Marshal(append(votes_list, new_vote))
		stub.PutState("Votes", votesJson)
		
		fmt.Println("Votes: " + string(votes))
		return shim.Success([]byte("Vote created"))
}

// Reply on a specific vote
func replyVote(stub shim.ChaincodeStubInterface, voteId string, reply string) peer.Response {	
	MSPid, _ := shim.GetMSPID()
	var OclToken int
	
	// Init vote receive list
	votes_list := []Votes{}
	Operator_list := []Operators{}

	// Get all votes
	votes, _ := stub.GetState("Votes")
	json.Unmarshal(votes, &votes_list)

	// Get Operator list
	value, _ := stub.GetState("Operators")
	json.Unmarshal(value, &Operator_list)

	i, _ := strconv.Atoi(voteId)

	for i := 0; i < len(Operator_list); i++ {
		if(Operator_list[i].OperatorID == MSPid) {
			OclToken = Operator_list[i].OclToken
		}
	}
	
	vote_Operator := Operators{OperatorID: MSPid, OclToken: OclToken}

	if(reply == "yes") {
		votes_list[i].Yes = append(votes_list[i].Yes, vote_Operator)
		votesJson, _ := json.Marshal(votes_list)
		stub.PutState("Votes", votesJson)
	} else if(reply == "no") {
		votes_list[i].No = append(votes_list[i].No, vote_Operator)
		votesJson, _ := json.Marshal(votes_list)
		stub.PutState("Votes", votesJson)
	}

	temp, _ := json.Marshal(votes_list[i])
	fmt.Println("tesstje: " + string(temp))
	
	return shim.Success([]byte("Replied on vote: " + votes_list[i].Title))
}

// Get all the current votes
func get(stub shim.ChaincodeStubInterface, voteType string) peer.Response {
	// Holds a string for the response
	var votes string

	// Local variables for value & error
	var get_value  []byte
	var err    error
	
	if get_value, err = stub.GetState(voteType); err != nil {

		fmt.Println("Get Failed!!! ", err.Error())

		return shim.Error(("Get Failed!! "+err.Error()+"!!!"))

	} 

	// nil indicates non existent key
	if (get_value == nil) {
		votes = "-1"
	} else {
		votes = string(get_value)
	}

	fmt.Println(votes)
	
	return shim.Success(get_value)
}

// String in array function
func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

// Close vote and 
func closeVotes(stub shim.ChaincodeStubInterface, Id string) peer.Response {
	// Min procent of the registerd operators that have to reply on
	// a vote.
	VOTE_MIN := 60
	PROP_PASS := 1.2

	// Init vote receive list
	var csl []string
	votes_list := []Votes{}
	Operator_list := []Operators{}
	

	index, _ := strconv.Atoi(Id)
	
	// Get all votes
	votes, _ := stub.GetState("Votes")
	json.Unmarshal(votes, &votes_list)

	// Get Operator list
	value, _ := stub.GetState("Operators")
	json.Unmarshal(value, &Operator_list)

	// Difference between time
	duration, _ := strconv.Atoi(votes_list[index].Duration)
	diff := time.Now().Sub(votes_list[index].Timestamp)

	fmt.Println("Test:")
	fmt.Println(diff)
	fmt.Println("Test2:")
	fmt.Println(time.Duration(duration)*time.Minute)

	if(diff > time.Duration(duration)*time.Minute) {
		Operator_count := len(Operator_list)
		reply_count := len(votes_list[index].Yes) + len(votes_list[index].No)
		reply_percentage := ((reply_count / Operator_count) * 100)
		

		if(reply_percentage >= VOTE_MIN) {
			var totalVotingPower float64
			var votingPowerYes float64
			var votingPowerNo float64 

			for i := 0; i < len(votes_list[index].No); i++ {
				votingPowerNo += float64(votes_list[index].No[i].OclToken)
				totalVotingPower += float64(votes_list[index].No[i].OclToken)

				// Append to check if operator voted
				csl = append(csl, votes_list[index].No[i].OperatorID)
			}

			for i := 0; i < len(votes_list[index].Yes); i++ {
				votingPowerYes += float64(votes_list[index].Yes[i].OclToken)
				totalVotingPower += float64(votes_list[index].Yes[i].OclToken)

				// Append to check if operator voted
				csl = append(csl, votes_list[index].Yes[i].OperatorID)
			}

			if(((votingPowerYes - votingPowerNo) / totalVotingPower) * 100 > PROP_PASS) {
				fmt.Println("Passed")
				votes_list[index].Answer = "Passed"

				for i := 0; i < len(Operator_list); i++ {
					if(stringInSlice(Operator_list[i].OperatorID, 
									 csl)) {
						Operator_list[i].OclToken += 2
					} else {
						Operator_list[i].OclToken -= 2
					}
				}

			} else {
				fmt.Println("Not passed")
				votes_list[index].Answer = "Not Passed"
			}
		} else {
			fmt.Println("test 2")
			return shim.Success([]byte("Cannot close, not enough replies")) 
		}
	} else {
		// Close not possible 
		return shim.Success([]byte("Close not possible, reply time no passed.")) 
	}

	// Close possible 
	voteJson, _ := json.Marshal(votes_list)
	stub.PutState("Votes", voteJson)

	OperatorJson, _ := json.Marshal(Operator_list)
	stub.PutState("Operators", OperatorJson)
	
	return shim.Success([]byte("Vote closed" + votes_list[index].Title)) 
}

func reset(stub shim.ChaincodeStubInterface) peer.Response {
	// Init lists
	voter_list, _ := json.Marshal([]Operators{})
	votes, _ := json.Marshal([]Votes{})
	closedVotes, _ := json.Marshal([]Votes{})

	// Put states to level DB
	stub.PutState("Operators", voter_list)
	stub.PutState("Votes", votes)
	stub.PutState("closedVotes", closedVotes)
	
	// Return success
	return shim.Success([]byte("Reset"))
}

// Chaincode registers with the Shim on startup
func main() {
	fmt.Printf("Started Chaincode. token/v10\n")
	err := shim.Start(new(Operators))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}