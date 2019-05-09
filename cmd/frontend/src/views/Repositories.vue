<template>
    <div>
        <base-header type="info" class="pb-6 pb-8 pt-5 pt-md-8"></base-header>

        <div class="container-fluid mt--7">
            
            <div class="card shadow"
                :class="type === 'dark' ? 'bg-default': ''">
                <div class="card-header border-0"
                    :class="type === 'dark' ? 'bg-transparent': ''">
                <div class="row align-items-center">
                    <div class="col">
                    <h3 class="mb-0" :class="type === 'dark' ? 'text-white': ''">
                        Repositories
                    </h3>
                    </div>
                    <div class="col text-right">
                    <base-button v-if="seeAll" @click="showAll" type="primary" size="sm">See all</base-button>
                    </div>
                </div>
                </div>

                <div class="table-responsive">
                    <base-table class="table align-items-center table-flush"
                                @rowClicked="goToRepo"
                                :class="type === 'dark' ? 'table-dark': ''"
                                :thead-classes="type === 'dark' ? 'thead-dark': 'thead-light'"
                                tbody-classes="list"
                                :data="reposToDisplay">
                        <template slot="columns">
                        <th>Repository Name</th>
                        <th>Project Name</th>
                        <th>Enabled</th>
                        <th>Owners</th>
                        <th></th>
                        </template>

                        <template slot-scope="{row}">
                        <th scope="row">
                            <div class="media align-items-center">
                            <!-- <a href="#" class="avatar rounded-circle mr-3">
                                <img alt="Image placeholder" :src="row.img">
                            </a> -->
                            <div class="media-body">
                                <span class="name mb-0 text-sm">{{row.name}}</span>
                            </div>
                            </div>
                        </th>
                        <td class="budget">
                            {{row.projectName}}
                        </td>

                        <td>
                            <badge class="badge-dot mr-4" :type="row.enabled ? 'success' : 'danger'">
                            <i :class="row.enabled ? 'bg-success' : 'bg-danger' "></i>
                            <span v-if="row.enabled" class="status">enabled</span>
                            <span v-else class="status">disabled</span>
                            </badge>
                        </td>

                        <td class="budget">
                            <div v-if="row.owners">
                                {{row.owners[0]}}
                            </div>
                            <div v-else>
                                No Owner
                            </div>

                        </td>

                        <td class="text-right">
                            <base-dropdown class="dropdown"
                                        position="right">
                            <a slot="title" class="btn btn-sm btn-icon-only text-light" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                                <i class="fas fa-ellipsis-v"></i>
                            </a>

                            <template>
                                <a class="dropdown-item" @click="goToRepo(row)">View/Edit</a>
                                
                                <a class="dropdown-item" href="#" @click="toggleEnableRepo(row)">
                                    <span v-if="row.enabled" class="status">Disable</span>
                                    <span v-else class="status">Enable</span>
                                </a>
                            </template>
                            </base-dropdown>
                        </td>

                        </template>

                    </base-table>
                </div>

                <div class="card-footer d-flex justify-content-end"
                    :class="type === 'dark' ? 'bg-transparent': ''">
                <base-pagination @input="changePage" :value=currentPage :perPage=reposPerPage :total=repositories.length></base-pagination>
                </div>

            </div>
        </div>

    </div>
</template>
<script>
  import ProjectsTable from './Tables/ProjectsTable'
  import axios from 'axios'
  export default {
    name: 'repositories',
    data: function () {
        return {
            repositories: [],
            reposToDisplay: [],
            type: "hello",
            reposPerPage: 5,
            currentPage: 1,

            //flags - this is probably not best practice but i'm going fast on this one
            seeAll: true
        }
    },
    components: {
      ProjectsTable
    },
    created(){
        axios.get('https://devopshelper.io/api/repositories', {
            headers: {
                'Authorization': 'Bearer ' + this.$store.state.user.token
            }
        })
        .then(response => this.repositories = response.data)
        .catch(error => console.log(error))
    },
    methods: {
        showAll(){
            this.currentPage = 1
            this.reposPerPage = this.repositories.length
            this.seeAll = false
            this.reposToDisplay = this.repositories
        },
        toggleEnableRepo(repo){
            repo.enabled = !repo.enabled
            //TODO: actually make the repo call
        },
        goToRepo(repo){
            this.$router.push({ name: 'repository', params: {repo: repo }})
        },
        changePage(pageNumber){
            let startIndex = this.reposPerPage * (pageNumber - 1)
            this.reposToDisplay = this.repositories.slice(startIndex, startIndex+this.reposPerPage)
            this.currentPage = pageNumber
        }
    }
  };
</script>
<style></style>
