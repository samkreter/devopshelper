set -e

###### Required
# export RESOURCE_GROUP
# export CLUSTER_NAME
# export SUBSCRIPTION_ID
# export COSMOSDB_ACCOUNT_NAME
# export COSMOSDB_DBNAME

az account set -s $SUBSCRIPTION_ID

########### Create an aks cluster ############
echo "Deploying AKS cluster: ${CLUSTER_NAME}....."
az aks create -g $RESOURCE_GROUP -n $CLUSTER_NAME --node-count 2

echo "Adding kubeconfig...."
az aks get-credentials -g $RESOURCE_GROUP -n $CLUSTER_NAME


########### Create an cosmosdb mongo ############
echo "Creating cosmosdb server: ${COSMOSDB_ACCOUNT_NAME}..."
az cosmosdb create \
    --resource-group $RESOURCE_GROUP \
    --name $COSMOSDB_ACCOUNT_NAME \
    --kind MongoDB 

echo "Creating cosmosdb database: ${COSMOSDB_DBNAME}...."
az cosmosdb database create \
    --resource-group $RESOURCE_GROUP \
    --name $COSMOSDB_ACCOUNT_NAME \
    --db-name $COSMOSDB_DBNAME

echo "Mongo Connection string saved to env var: MONGO_CONNECTION_STRING"
export MONGO_CONNECTION_STRING=$(az cosmosdb list-connection-strings -g ${RESOURCE_GROUP} -n ${COSMOSDB_ACCOUNT_NAME} --query 'connectionStrings[0].connectionString')

echo "Successfully provisioned base infrastructure"