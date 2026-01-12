const inputTel = document.getElementById('inputTel');

inputTel.addEventListener('input', function (e) {
  let value = e.target.value;
  value = value.replace(/\D/g, ''); // Удаляем все, кроме цифр

  // Добавляем символы "+7 (", ")" и "-" последовательно
  let formattedValue = '';
  if (value.length > 0) {
    formattedValue = '+7';
  }
  if (value.length >= 2) {
    formattedValue += ' (' + value.slice(1, 4);
  }
  if (value.length >= 5) {
    formattedValue += ') ' + value.slice(4, 7);
  }
  if (value.length >= 8) {
    formattedValue += '-' + value.slice(7, 9);
  }
  if (value.length >= 10) {
    formattedValue += '-' + value.slice(9, 11);
  }

  e.target.value = formattedValue;
});

async function get(url) {
  return fetch(url, {
    method: 'GET',
  })
    .then(response => response.json())
    .then(data => {
      console.log(url);
      console.log(data);
      populateSpecialties(data.specialties);
    })
    .catch(error => {
      console.error('Ошибка', error);
    });
}
const url = `https://mis-api.kreosoft.space/api/dictionary/speciality?size=30`;
get(url)

function populateSpecialties(specialties) {
  const selectSpecialties = document.getElementById('selectSpecialties');
  specialties.forEach(specialty => {
    const option = document.createElement('option');
    option.value = specialty.id;
    option.text = specialty.name;
    selectSpecialties.appendChild(option);
  });
}


async function registerPost(data) {
  const url = 'https://mis-api.kreosoft.space/api/doctor/register';
  return fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      console.log(result);
      const errorMessage = document.getElementById('errorMessage');
      errorMessage.textContent = '';

      if (result.message) {
        errorMessage.textContent = result.message;
        console.log(result.message);
      }

      if (result.title) {
        errorMessage.textContent = result.title;
        console.log(result.title);
      }

      if (result.token) {
        token = result.token;
        localStorage.setItem('token', token);
        window.location.href = '../patients/patients.html';
      }
    })
    .catch(error => {
      console.error('Ошибка', error);
      const errorMessage = document.getElementById('errorMessage');
      errorMessage.textContent = 'Произошла ошибка при регистрации. Пожалуйста, попробуйте еще раз.';
    });
}

const form = document.querySelector('form');
if (form) {
  form.addEventListener('submit', function (event) {
    event.preventDefault(); // Предотвращаем отправку формы

    const name = document.getElementById('inputName').value;
    const password = document.getElementById('inputPassword').value;
    const email = document.getElementById('Email').value;
    const birthday = document.getElementById('inputData').value;
    const gender = document.getElementById('selectGender').value;
    const phone = document.getElementById('inputTel').value;
    const speciality = document.getElementById('selectSpecialties').value;

    const data = {
      name: name,
      password: password,
      email: email,
      birthday: birthday,
      gender: gender,
      phone: phone,
      speciality: speciality
    };

    if (!phone) {
      data.phone = null;
    }

    console.log(data);

    registerPost(data);
  });
}

