<template>
    <div>
        <base-header type="info" class="pb-6 pb-8 pt-5 pt-md-8">
            <!-- <div class="row">
                <div class="col-xl-3 col-lg-6">
                    <stats-card title="Total traffic"
                                type="gradient-red"
                                sub-title="350,897"
                                icon="ni ni-active-40"
                                class="mb-4 mb-xl-0">

                        <template slot="footer">
                            <span class="text-success mr-2"><i class="fa fa-arrow-up"></i> 3.48%</span>
                            <span class="text-nowrap">Since last month</span>
                        </template>
                    </stats-card>
                </div>
                <div class="col-xl-3 col-lg-6">
                    <stats-card title="Total traffic"
                                type="gradient-orange"
                                sub-title="2,356"
                                icon="ni ni-chart-pie-35"
                                class="mb-4 mb-xl-0">

                        <template slot="footer">
                            <span class="text-success mr-2"><i class="fa fa-arrow-up"></i> 12.18%</span>
                            <span class="text-nowrap">Since last month</span>
                        </template>
                    </stats-card>
                </div>
                <div class="col-xl-3 col-lg-6">
                    <stats-card title="Sales"
                                type="gradient-green"
                                sub-title="924"
                                icon="ni ni-money-coins"
                                class="mb-4 mb-xl-0">

                        <template slot="footer">
                            <span class="text-danger mr-2"><i class="fa fa-arrow-down"></i> 5.72%</span>
                            <span class="text-nowrap">Since last month</span>
                        </template>
                    </stats-card>

                </div>
                <div class="col-xl-3 col-lg-6">
                    <stats-card title="Performance"
                                type="gradient-info"
                                sub-title="49,65%"
                                icon="ni ni-chart-bar-32"
                                class="mb-4 mb-xl-0">

                        <template slot="footer">
                            <span class="text-success mr-2"><i class="fa fa-arrow-up"></i> 54.8%</span>
                            <span class="text-nowrap">Since last month</span>
                        </template>
                    </stats-card>
                </div>
            </div> -->
        </base-header>

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
                    <base-button type="primary" size="sm">See all</base-button>
                    </div>
                </div>
                </div>

                <div class="table-responsive">
                <base-table class="table align-items-center table-flush"
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
                            <a class="dropdown-item" href="#">Action</a>
                            <a class="dropdown-item" href="#">Another action</a>
                            <a class="dropdown-item" href="#">Something else here</a>
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
        }
    },
    components: {
      ProjectsTable
    },
    created(){
        axios.get('https://devopshelper.eastus.cloudapp.azure.com/api/repositories', {
            headers: {
                'Authorization': 'Bearer ' + this.$store.state.user.token
            }
        })
        .then(response => this.repositories = response.data)
        .catch(error => console.log(error))
    },
    methods: {
        changePage(pageNumber){
            let startIndex = this.reposPerPage * (pageNumber - 1)
            this.reposToDisplay = this.repositories.slice(startIndex, startIndex+5)
            this.currentPage = pageNumber
        }
    }
  };
</script>
<style></style>
