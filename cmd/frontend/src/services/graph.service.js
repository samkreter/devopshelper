import axios from 'axios'

export default class GraphService {
    constructor() {
      this.graphUrl = 'https://graph.microsoft.com/v1.0/';
    }
  
    getUserInfo(token) {
      const headers = new Headers({ Authorization: `Bearer ${token}` });
      const options = {
        headers
      };
      return fetch(`${this.graphUrl}/me`, options)
        .then(response => response.json())
        .catch(response => {
          throw new Error(response.text());
        });
    }

    getUserPhoto(token) {
      const headers = new Headers({ Authorization: `Bearer ${token}` });
      const options = {
        headers
      };

      return axios(`${this.graphUrl}me/photo/$value`, { headers: { Authorization: `Bearer ${token}` }, responseType: 'arraybuffer' })
        .then(response => {
              if (response.status == 200) {
                const avatar = "data:" + response.headers["content-type"] + ";base64," + new Buffer(response.data, 'binary').toString('base64');    
                return avatar
              }
              return ""
              
            })
            .catch(err => {
                console.log(err)
            });
    }
  }