<template>
    <div id="login">
        <button type="button" v-on:click="login()">Login</button>
    </div>
</template>

<script>
    import AuthService from '../services/auth.service';
    import GraphService from '../services/graph.service';

    export default {
        name: 'Login',
        data() {
            return {
                user: null,
                userInfo: null,
                apiCallFailed: false,
                loginFailed: false
            }
        },
        created() {
          this.authService = new AuthService();
          this.graphService = new GraphService();
        },
        methods: {
            login() {
              this.loginFailed = false;
              this.authService.login().then(
                user => {
                  if (user) {
                    this.user = user;
                    this.callAPI()
                  } else {
                    this.loginFailed = true;
                  }
                },
                () => {
                  this.loginFailed = true;
                }
              );
            },
            callAPI() {
                this.apiCallFailed = false;
                this.authService.getToken().then(
                    token => {
                    this.graphService.getUserInfo(token).then(
                        data => {
                        this.userInfo = data;
                        },
                        error => {
                        console.error(error);
                        this.apiCallFailed = true;
                        }
                    );
                    },
                    error => {
                    console.error(error);
                    this.apiCallFailed = true;
                    }
                );
            }
        }
    }
</script>
