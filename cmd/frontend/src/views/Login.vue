<template>
        <div>
            <base-header class="header pb-8 pt-5 pt-lg-8 d-flex align-items-center"
                        style="min-height: 600px; background-size: cover; background-position: center top;">
                <!-- Mask -->
                <span class="mask bg-primary"></span>
                <!-- Header container -->
                <div class="container-fluid d-flex align-items-center">
                    <div class="row justify-content-center">
                        <div class="col-lg-7 col-md-10">
                            <h1 class="display-2 text-white">Welcome to Devops Helper!</h1>
                            <p class="text-white mt-0 mb-5">Currently only available for Microsoft Employees.</p>
                            <p class="text-white display-5" >Sign In With</p>
                            <a v-on:click="login()" class="btn btn-neutral btn-icon">
                                <span class="btn-inner--icon"><img src="img/icons/common/microsoft.jpg"></span>
                                <span class="btn-inner--text">Microsoft</span>
                            </a>
                        </div>
                    </div>
                </div>
            </base-header>
        </div>
</template>

<script>
    import GraphService from '../services/graph.service';

    export default {
        name: 'Login',
        created() {
          this.graphService = new GraphService();
            if (this.$store.state.user){
                this.$router.push({ name: "repositories" });
            }
        },
        methods: {
            login() {
              this.$authService.login().then(
                user => {
                  if (user) {
                    this.$authService.getToken().then(
                        token => {
                            user.token = token
                            this.$store.commit("setUser", user)
                            this.$router.push({ name: "repositories" });
                            // this.graphService.getUserInfo(token).then(
                            //     data => {
                            //         this.$emit("authenticated", data);
                            //         this.$router.replace({ name: "repositories" });
                            //     },
                            //     error => {
                            //     console.error(error);
                            //     }
                            // );
                        },
                        error => {
                            console.error(error);
                        }
                    );
                  } else {
                    console.log("Login failed")
                  }
                },
                () => {
                  console.log("Login failed")
                }
              );
            }
        }
    }
</script>

<style>
</style>
