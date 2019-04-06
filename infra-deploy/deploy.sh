export RESOURCE_GROUP=<myresourcegroup>
export CLUSTER_NAME=<myclusterName>

## Create an aks cluster
as aks create -g $RESOURCE_GROUP -n $CLUSTER_NAME


## Add dns for the IP
IP="Ingress IP"
# Name to associate with public IP address
DNSNAME="devops-reviewer"
# Get the resource-id of the public ip
PUBLICIPID=$(az network public-ip list --query "[?ipAddress!=null]|[?contains(ipAddress, '$IP')].[id]" --output tsv)
# Update public ip address with DNS name
az network public-ip update --ids $PUBLICIPID --dns-name $DNSNAME

