export RESOURCE_GROUP=<myresourcegroup>
export CLUSTER_NAME=<myclusterName>

########### Create an aks cluster ############
#as aks create -g $RESOURCE_GROUP -n $CLUSTER_NAME



########### Creating TLS Ingress Controller ############
helm install stable/nginx-ingress --namespace kube-system --set controller.replicaCount=2



########### Add dns for the IP ##########
IP="<get Ingress IP Address>"
# Name to associate with public IP address
DNSNAME="devops-reviewer"
# Get the resource-id of the public ip
PUBLICIPID=$(az network public-ip list --query "[?ipAddress!=null]|[?contains(ipAddress, '$IP')].[id]" --output tsv)
# Update public ip address with DNS name
az network public-ip update --ids $PUBLICIPID --dns-name $DNSNAME




########### Install Cert Manager ################
kubectl label namespace kube-system certmanager.k8s.io/disable-validation=true

kubectl apply \
    -f https://raw.githubusercontent.com/jetstack/cert-manager/release-0.6/deploy/manifests/00-crds.yaml

helm install stable/cert-manager \
    --namespace kube-system \
    --set ingressShim.defaultIssuerName=letsencrypt-staging \
    --set ingressShim.defaultIssuerKind=ClusterIssuer \
    --version v0.6.6

## Create the cluster issuer resource
kubectl apply -f cluster-issuer.yaml

## Create a certificate resource 
kubectl apply -f certificates.yaml



