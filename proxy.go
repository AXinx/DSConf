// This chaincode calls the token/v5
// This is to JUST demonstrate the invoke mechanism
// This cc will act as a proxy
package main

import (
	"fmt"

	// April 2020, Updated for Fabric 2.0
	// Video may have shim package import for Fabric 1.4 - please ignore

	"github.com/hyperledger/fabric-chaincode-go/shim"
	
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

// CallerChaincode Represents our chaincode object
type CallerChaincode struct {
}

// Channel Name
const    Channel = "airlinechannel"

// Init func will do nothing
func (token *CallerChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Init executed.")
	// Return success
	return shim.Success([]byte("Init Done."))
}

// Invoke method
func (token *CallerChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	var TargetChaincode string
	funcName, args := stub.GetFunctionAndParameters()

	// Chaincode to be invoked
	if(funcName == "vote"){
		TargetChaincode = "token"
	} else if(funcName == "action") {
		TargetChaincode = "auto_action"
	} else {
		// This is not good
		return shim.Error(("Bad Function Name from caller = " + funcName + "!!!"))
	}

	args_output := make([][]byte, len(args))
	for i, v := range args {
		args_output[i] = []byte(v)
	}

	// Sets the value of MyToken in token chaincode (V5)
	response := stub.InvokeChaincode(TargetChaincode, args_output, Channel)

	return shim.Success([]byte(response.String()))
}

// Chaincode registers with the Shimtrue on startup
func main() {
	fmt.Printf("Started Chaincode. caller/v10\n")
	err := shim.Start(new(CallerChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}