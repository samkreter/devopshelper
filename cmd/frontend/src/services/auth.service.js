import * as Msal from 'msal';

export default class AuthService {
  constructor() {
    // //let PROD_REDIRECT_URI = 'https://sunilbandla.github.io/vue-msal-sample/';
    // let redirectUri = window.location.origin;
    // // if (window.location.hostname !== '127.0.0.1') {
    // //   redirectUri = PROD_REDIRECT_URI;
    // // }
    this.applicationConfig = {
      clientID: '7bdce0c4-74d3-4dca-89cd-9ca67ec69eab',
      graphScopes: ["user.read"],
      authority: "https://login.microsoftonline.com/72f988bf-86f1-41af-91ab-2d7cd011db47"
    };

    if (window.location.hash.includes("id_token")) {
      new Msal.UserAgentApplication("4ecf3d26-e844-4855-9158-b8f6c0121b50", null, null);
    }

    this.app = new Msal.UserAgentApplication(
      this.applicationConfig.clientID,
      this.applicationConfig.authority,
      this.acquireTokenRedirectCallBack,
      {storeAuthStateInCookie: true,
        //redirectUri: "msal7bdce0c4-74d3-4dca-89cd-9ca67ec69eab://auth",
        cacheLocation: "localStorage",
        navigateToLoginRequestUrl: false}
    );
  }
  login() {
    return this.app.loginPopup(this.applicationConfig.graphScopes).then(
      idToken => {
        const user = this.app.getUser();
        if (user) {
          return user;
        } else {
          return null;
        }
      },
      () => {
        return null;
      }
    );
  };
  logout() {
    this.app.logout();
  };
  getToken() {
    return this.app.acquireTokenSilent(this.applicationConfig.graphScopes).then(
      accessToken => {
        return accessToken;
      },
      error => {
        return this.app
          .acquireTokenPopup(this.applicationConfig.graphScopes)
          .then(
            accessToken => {
              return accessToken;
            },
            err => {
              console.error(err);
            }
          );
      }
    );
  };
  acquireTokenRedirectCallBack(errorDesc, token, error, tokenType)
  {
   if(tokenType === "access_token")
   {
       //callMSGraph(applicationConfig.graphEndpoint, token, graphAPICallback);
   } else {
          console.log("token type is:"+tokenType);
   } 
  }
}
