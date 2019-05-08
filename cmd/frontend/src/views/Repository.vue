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
                        {{ this.repo.name }}
                    </h3>
                    </div>
                    <div class="col text-right">
                    <base-button type="primary" size="sm">See all</base-button>
                    </div>
                </div>
                </div>

                <!-- ####### Core Reviewers  -->
                <div class="table-responsive">
                <base-table class="table align-items-center table-flush"
                            :class="type === 'dark' ? 'table-dark': ''"
                            :thead-classes="type === 'dark' ? 'thead-dark': 'thead-light'"
                            tbody-classes="list"
                            :data="coreReviewersToDisplay">
                    <template slot="columns">
                    <th>Reviewer Group</th>
                    <th>Alias</th>
                    <th>Required</th>
                    <th></th>
                    </template>

                    <template slot-scope="{row}">
                    <th scope="row">
                        <div class="media align-items-center">
                        <!-- <a href="#" class="avatar rounded-circle mr-3">
                            <img alt="Image placeholder" :src="row.img">
                        </a> -->
                        <div class="media-body">
                            <span class="name mb-0 text-sm">Core</span>
                        </div>
                        </div>
                    </th>
                    <td class="budget">
                        {{row.alias}}
                    </td>

                    <td>
                        <badge class="badge-dot mr-4" :type="coreGroup.required ? 'success' : 'danger'">
                        <i :class="coreGroup.required ? 'bg-success' : 'bg-danger' "></i>
                        <span v-if="coreGroup.required" class="status">required</span>
                        <span v-else class="status">not required</span>
                        </badge>
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
                <base-pagination @input="changeCorePage" :value=coreCurrentPage :perPage=reviewersPerPage :total=coreReviewers.length></base-pagination>
                </div>

                <!-- ####### Core Reviewers  -->
                
                <!-- ####### Secondary Reviewers  -->
                <div class="table-responsive">
                <base-table class="table align-items-center table-flush"
                            :class="type === 'dark' ? 'table-dark': ''"
                            :thead-classes="type === 'dark' ? 'thead-dark': 'thead-light'"
                            tbody-classes="list"
                            :data="secondReviewersToDisplay">
                    <template slot="columns">
                    <th>Reviewer Group</th>
                    <th>Alias</th>
                    <th>Required</th>
                    <th></th>
                    </template>

                    <template slot-scope="{row}">
                    <th scope="row">
                        <div class="media align-items-center">
                        <!-- <a href="#" class="avatar rounded-circle mr-3">
                            <img alt="Image placeholder" :src="row.img">
                        </a> -->
                        <div class="media-body">
                            <span class="name mb-0 text-sm">Additional</span>
                        </div>
                        </div>
                    </th>
                    <td class="budget">
                        {{row.alias}}
                    </td>

                    <td>
                        <badge class="badge-dot mr-4" :type="secondGroup.required ? 'success' : 'danger'">
                        <i :class="secondGroup.required ? 'bg-success' : 'bg-danger' "></i>
                        <span v-if="secondGroup.required" class="status">required</span>
                        <span v-else class="status">not required</span>
                        </badge>
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
                <base-pagination @input="changeSecondPage" :value=secondCurrentPage :perPage=reviewersPerPage :total=secondReviewers.length></base-pagination>
                </div>

                <!-- ####### Secondary Reviewers  -->

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
            repository: {},
            reviewersPerPage: 5,
            
            // Core Reveiwers
            coreGroup: {},
            coreReviewersToDisplay: [],
            coreReviewers: [],
            coreCurrentPage: 1,

            // Secondary Reveiwers
            secondGroup: {},
            secondReviewersToDisplay: [],
            secondReviewers: [],
            secondCurrentPage: 1,
            
            type: "hello"
        }
    },
    created(){
        this.coreReviewers = this.repo.reviewerGroups.Tier1.reviewers
        this.coreCurrentPage = 1
        this.changeCorePage(1)

        this.secondReviewers = this.repo.reviewerGroups.Tier2.reviewers
        this.secondCurrentPage = 1
        this.changeSecondPage(1)

        this.coreGroup = this.repo.reviewerGroups.Tier1
        this.secondGroup = this.repo.reviewerGroups.Tier2
    },
    methods: {
        changeCorePage(pageNumber){
            let startIndex = this.reviewersPerPage * (pageNumber - 1)
            this.coreReviewersToDisplay = this.coreReviewers.slice(startIndex, startIndex+5)
            this.coreCurrentPage = pageNumber
        },
        changeSecondPage(pageNumber){
            let startIndex = this.reviewersPerPage * (pageNumber - 1)
            this.secondReviewersToDisplay = this.secondReviewers.slice(startIndex, startIndex+5)
            this.secondCurrentPage = pageNumber
        }
    },
    props: ["repo"]
  };
</script>
<style></style>
