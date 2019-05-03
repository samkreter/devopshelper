import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

export const store = new Vuex.Store({
    state: {
        user:  getUserFromLocal()
    },
    mutations: {
        setUser(state, user){
            user.storedTime = new Date().getTime();
            localStorage.setItem("currUser", JSON.stringify(user))
            state.user = user
        },
        deleteUser(state) {
            localStorage.removeItem("currUser")
            state.user = null
        }
    }
})

function getUserFromLocal(){
    let userStr = localStorage.getItem("currUser")
    if (!userStr){
        return null
    }

    let user = JSON.parse(userStr)

    // Timeout if token has been in store longer than 30 mintues
    if ((new Date()) - new Date(user.storedTime) > (30 * 60000)) {
        localStorage.removeItem("currUser")
        return null
    }

    return user
}