# VSTSAutoReviewer

The autoReveiwer will add the functionallity to assign reviewers to open PRs in Visual Studios Team Services.

Different reviewer groups and wheather the reviewer is optional is specificed through a json file. The state of next reviewer to be assigned is also saved. 

## Using Azure Container Instance and Azure Logic Apps

1. Create an Azure Storage Account using either the portal or the CLI. 

2. Add the reviewers.json file to the file share. An example of the format can be found in the examples folder.

3. Either build and push the image to a container registry or the latest stable version can be pulled from the dockerhub repository pskreter/vstsreviewer:stable

4. Create an Azure Logic App with a timed trigger of 10 minutes (or whatever you think is good for your team.)

5. Next add the ACI create component. Attach the fileshare as a volume at /config and add the location of the pushed image.

Now you should be good to start running. 

