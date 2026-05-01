function getDaily() {
  fetch('http://localhost:8080/api/game/daily')
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