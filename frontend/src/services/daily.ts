const BACKEND_URL = import.meta.env.VITE_BACKEND_URL

function getDaily() {

  fetch(`${BACKEND_URL}/api/game/daily`)
    .then(response => {
      if (!response.ok) {
        throw new Error('Request failed');
      }
      return response.json();
    })
    .then(data => console.log(data))
    .catch(error => console.error(error));
}

function doAttempt() {};

export {getDaily, doAttempt}