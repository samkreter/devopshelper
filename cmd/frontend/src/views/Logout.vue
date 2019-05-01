<template>
</template>

<script>
export default {
    name: 'logout',
    created(){
        this.$store.commit("deleteUser")
        this.$router.push({name: "login"})
    }
}
</script>

<style>

</style>
