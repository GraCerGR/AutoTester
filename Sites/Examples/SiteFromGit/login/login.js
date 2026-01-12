let token;

async function post(url, data = null) {
  return fetch(url, {
    method: 'POST',
    body: JSON.stringify(data),
    headers: new Headers({
      'Content-Type': 'application/json'
    })
  })
    .then(response => response.json())
    .then(data => {
      console.log(data);
      if (data.token) {
        token = data.token;
        localStorage.setItem('token', token);
        window.location.href = '../patients/patients.html';
      } else {
        alert(data.message || 'Неверный логин или пароль');
      }
      localStorage.setItem('token', token);
    })
    .catch(error => {
      console.error('Ошибка', error);
    });
}
const url = "https://mis-api.kreosoft.space/api/doctor/login";
const form = document.querySelector('form');
if (form) {
  form.addEventListener('submit', function (event) {
    event.preventDefault(); // Предотвращаем отправку формы

    const email = document.getElementById('Email').value;
    const password = document.getElementById('password').value;
    const data = {
      email: email,
      password: password
    };

    console.log(data);
    post(url, data);
    console.log(token);
  });
}
