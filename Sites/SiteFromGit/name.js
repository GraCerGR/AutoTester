var token = localStorage.getItem('token');

async function getName(url, token) {
  return fetch(url, {
    method: 'GET',
    headers: new Headers({
      "Authorization": `Bearer ${token}`
    }),
  })
    .then(response => response.json())
    .then(data => {
      document.getElementById('inputNameEntry').textContent = data.name;
    })
    .catch(error => {
      console.error('Ошибка', error);
    });
}
getName(`https://mis-api.kreosoft.space/api/doctor/profile`, token)