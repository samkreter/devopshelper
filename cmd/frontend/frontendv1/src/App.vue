<template>
  <div id="app">
    <notifications></notifications>
    <router-view @authenticated="setLoginUser" />
  </div>
</template>

<script>
    export default {
        name: 'App',
        data() {
            return {
                user: null,
            }
        },
        mounted() {
            if(!this.user) {
                this.$router.replace({ name: "login" });
            }
        },
        methods: {
            setLoginUser(user) {
                console.log("getuser: ", user)
                this.user = user;
            },
            logout() {
                this.user = null;
            }
        }
    }
</script>
