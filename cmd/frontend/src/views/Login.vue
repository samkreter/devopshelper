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
