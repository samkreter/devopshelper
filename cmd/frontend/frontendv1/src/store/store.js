import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

export const store = new Vuex.Store({
    state: {
        user:  getUserFromLocal()
    },
    mutations: {
        setUser(state, user){
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

    return JSON.parse(userStr)
}