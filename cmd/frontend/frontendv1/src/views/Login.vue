<template>
        <div class="row justify-content-center">
            <div class="col-lg-5 col-md-7">
                <div class="card bg-secondary shadow border-0">
                    <div class="card-header bg-transparent pb-5">
                        <div class="text-muted text-center mt-2 mb-3"><small>Sign in with</small></div>
                        <div class="btn-wrapper text-center">
                            <a v-on:click="login()" class="btn btn-neutral btn-icon">
                                <span class="btn-inner--icon"><img src="img/icons/common/microsoft.jpg"></span>
                                <span class="btn-inner--text">Microsoft</span>
                            </a>
                        </div>
                    </div>
                </div>
            </div>
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
                    console.log("#######: ", user)
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

<style>
</style>
