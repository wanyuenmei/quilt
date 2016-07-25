package provider

import (
	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

// This is simply a wrapper around all the Azure calls. This makes it very easy to mock
// during testing or as needed.
type azureAPI interface {
	ifaceCreate(rgName string, ifaceName string, param network.Interface,
		cancel <-chan struct{}) (result autorest.Response, err error)
	ifaceDelete(rgName string, ifaceName string, cancel <-chan struct{}) (
		result autorest.Response, err error)
	ifaceGet(rgName string, ifaceName string, expand string) (
		result network.Interface, err error)

	publicIPCreate(rgName string, pipAddrName string, param network.PublicIPAddress,
		cancel <-chan struct{}) (result autorest.Response, err error)
	publicIPDelete(rgName string, pipAddrName string, cancel <-chan struct{}) (
		result autorest.Response, err error)
	publicIPGet(rgName string, pipAddrName string, expand string) (
		result network.PublicIPAddress, err error)

	securityGroupCreate(rgName string, nsgName string, param network.SecurityGroup,
		cancel <-chan struct{}) (result autorest.Response, err error)
	securityGroupList(rgName string) (result network.SecurityGroupListResult,
		err error)

	securityRuleCreate(rgName string, nsgName string, ruleName string,
		ruleParam network.SecurityRule, cancel <-chan struct{}) (
		result autorest.Response, err error)
	securityRuleDelete(rgName string, nsgName string, ruleName string,
		cancel <-chan struct{}) (result autorest.Response, err error)
	securityRuleList(rgName string, nsgName string) (
		result network.SecurityRuleListResult, err error)

	vnetCreate(rgName string, vnName string, param network.VirtualNetwork,
		cancel <-chan struct{}) (result autorest.Response, err error)
	vnetList(rgName string) (result network.VirtualNetworkListResult, err error)

	rgCreate(rgName string, param resources.ResourceGroup) (
		result resources.ResourceGroup, err error)
	rgDelete(rgName string, cancel <-chan struct{}) (result autorest.Response,
		err error)

	storageListByRg(rgName string) (result storage.AccountListResult, err error)
	storageCheckName(accountName storage.AccountCheckNameAvailabilityParameters) (
		result storage.CheckNameAvailabilityResult, err error)
	storageCreate(rgName string, accountName string,
		param storage.AccountCreateParameters, cancel <-chan struct{}) (
		result autorest.Response, err error)
	storageGet(rgName string, accountName string) (result storage.Account, err error)

	vmCreate(rgName string, vmName string, param compute.VirtualMachine,
		cancel <-chan struct{}) (result autorest.Response, err error)
	vmDelete(rgName string, vmName string, cancel <-chan struct{}) (
		result autorest.Response, err error)
	vmList(rgName string) (result compute.VirtualMachineListResult, err error)
}

type azureClient struct {
	ifaceClient    network.InterfacesClient
	publicIPClient network.PublicIPAddressesClient
	secGroupClient network.SecurityGroupsClient
	secRulesClient network.SecurityRulesClient
	vnetClient     network.VirtualNetworksClient
	rgClient       resources.GroupsClient
	storageClient  storage.AccountsClient
	vmClient       compute.VirtualMachinesClient
}

func newAzureClient(subscriptionID string, spt *azure.ServicePrincipalToken) azureAPI {
	client := azureClient{}

	client.ifaceClient = network.NewInterfacesClient(subscriptionID)
	client.ifaceClient.Authorizer = spt

	client.publicIPClient = network.NewPublicIPAddressesClient(subscriptionID)
	client.publicIPClient.Authorizer = spt

	client.secGroupClient = network.NewSecurityGroupsClient(subscriptionID)
	client.secGroupClient.Authorizer = spt

	client.secRulesClient = network.NewSecurityRulesClient(subscriptionID)
	client.secRulesClient.Authorizer = spt

	client.vnetClient = network.NewVirtualNetworksClient(subscriptionID)
	client.vnetClient.Authorizer = spt

	client.rgClient = resources.NewGroupsClient(subscriptionID)
	client.rgClient.Authorizer = spt

	client.storageClient = storage.NewAccountsClient(subscriptionID)
	client.storageClient.Authorizer = spt

	client.vmClient = compute.NewVirtualMachinesClient(subscriptionID)
	client.vmClient.Authorizer = spt

	return client
}

func (client azureClient) ifaceCreate(rgName string, ifaceName string,
	param network.Interface, cancel <-chan struct{}) (result autorest.Response,
	err error) {
	return client.ifaceClient.CreateOrUpdate(rgName, ifaceName, param, cancel)
}

func (client azureClient) ifaceDelete(rgName string, ifaceName string,
	cancel <-chan struct{}) (result autorest.Response, err error) {
	return client.ifaceClient.Delete(rgName, ifaceName, cancel)
}

func (client azureClient) ifaceGet(rgName string, ifaceName string, expand string) (
	result network.Interface, err error) {
	return client.ifaceClient.Get(rgName, ifaceName, expand)
}

func (client azureClient) publicIPCreate(rgName string, pipAddrName string,
	param network.PublicIPAddress, cancel <-chan struct{}) (result autorest.Response,
	err error) {
	return client.publicIPClient.CreateOrUpdate(rgName, pipAddrName, param, cancel)
}

func (client azureClient) publicIPDelete(rgName string, pipAddrName string,
	cancel <-chan struct{}) (result autorest.Response, err error) {
	return client.publicIPClient.Delete(rgName, pipAddrName, cancel)
}

func (client azureClient) publicIPGet(rgName string, pipAddrName string, expand string) (
	result network.PublicIPAddress, err error) {
	return client.publicIPClient.Get(rgName, pipAddrName, expand)
}

func (client azureClient) securityGroupCreate(rgName string, nsgName string,
	param network.SecurityGroup, cancel <-chan struct{}) (result autorest.Response,
	err error) {
	return client.secGroupClient.CreateOrUpdate(rgName, nsgName, param, cancel)
}

func (client azureClient) securityGroupList(rgName string) (
	result network.SecurityGroupListResult, err error) {
	return client.secGroupClient.List(rgName)
}

func (client azureClient) securityRuleCreate(rgName string, nsgName string,
	ruleName string, ruleParam network.SecurityRule, cancel <-chan struct{}) (
	result autorest.Response, err error) {
	return client.secRulesClient.CreateOrUpdate(rgName, nsgName, ruleName,
		ruleParam, cancel)
}

func (client azureClient) securityRuleDelete(rgName string, nsgName string,
	ruleName string, cancel <-chan struct{}) (result autorest.Response, err error) {
	return client.secRulesClient.Delete(rgName, nsgName, ruleName, cancel)
}

func (client azureClient) securityRuleList(rgName string, nsgName string) (
	result network.SecurityRuleListResult, err error) {
	return client.secRulesClient.List(rgName, nsgName)
}

func (client azureClient) vnetCreate(rgName string, vnName string,
	param network.VirtualNetwork, cancel <-chan struct{}) (result autorest.Response,
	err error) {
	return client.vnetClient.CreateOrUpdate(rgName, vnName, param, cancel)
}

func (client azureClient) vnetList(rgName string) (
	result network.VirtualNetworkListResult, err error) {
	return client.vnetClient.List(rgName)
}

func (client azureClient) rgCreate(rgName string, param resources.ResourceGroup) (
	result resources.ResourceGroup, err error) {
	return client.rgClient.CreateOrUpdate(rgName, param)
}

func (client azureClient) rgDelete(rgName string, cancel <-chan struct{}) (
	result autorest.Response, err error) {
	return client.rgClient.Delete(rgName, cancel)
}

func (client azureClient) storageListByRg(rgName string) (
	result storage.AccountListResult, err error) {
	return client.storageClient.ListByResourceGroup(rgName)
}

func (client azureClient) storageCheckName(
	accountName storage.AccountCheckNameAvailabilityParameters) (
	result storage.CheckNameAvailabilityResult, err error) {
	return client.storageClient.CheckNameAvailability(accountName)
}

func (client azureClient) storageCreate(rgName string, accountName string,
	param storage.AccountCreateParameters, cancel <-chan struct{}) (
	result autorest.Response, err error) {
	return client.storageClient.Create(rgName, accountName, param, cancel)
}

func (client azureClient) storageGet(rgName string, accountName string) (
	result storage.Account, err error) {
	return client.storageClient.GetProperties(rgName, accountName)
}

func (client azureClient) vmCreate(rgName string, vmName string,
	param compute.VirtualMachine, cancel <-chan struct{}) (result autorest.Response,
	err error) {
	return client.vmClient.CreateOrUpdate(rgName, vmName, param, cancel)
}

func (client azureClient) vmDelete(rgName string, vmName string,
	cancel <-chan struct{}) (result autorest.Response, err error) {
	return client.vmClient.Delete(rgName, vmName, cancel)
}

func (client azureClient) vmList(rgName string) (result compute.VirtualMachineListResult,
	err error) {
	return client.vmClient.List(rgName)
}
