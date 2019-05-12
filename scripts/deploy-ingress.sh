set -e


########### Add dns for the IP ##########
## REQUIRED VARS
# IP
# DNSNAME


# Get the resource-id of the public ip
PUBLICIPID=$(az network public-ip list --query "[?ipAddress!=null]|[?contains(ipAddress, '$IP')].[id]" --output tsv)
# Update public ip address with DNS name
echo "Updating public IP with DNS name: ${DNSNAME}"
az network public-ip update --ids $PUBLICIPID --dns-name $DNSNAME


########### Install Cert Manager ################
kubectl label namespace kube-system certmanager.k8s.io/disable-validation=true

kubectl apply \
    -f https://raw.githubusercontent.com/jetstack/cert-manager/release-0.6/deploy/manifests/00-crds.yaml

helm install stable/cert-manager \
    --name cert-manager \
    --namespace kube-system \
    --set ingressShim.defaultIssuerName=letsencrypt-staging \
    --set ingressShim.defaultIssuerKind=ClusterIssuer \
    --version v0.6.6

## NOTE: fill in fqdn in the certificates.yaml file and email in cluster-issuer
## Create the cluster issuer resource
kubectl apply -f ./deployFiles/cluster-issuer-prod.yaml

## Create a certificate resource 
kubectl apply -f ./deployFiles/certificate-prod.yaml