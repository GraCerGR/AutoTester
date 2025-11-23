var token = localStorage.getItem('token');
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

async function get(url, token) {
  return fetch(url, {
    method: 'GET',
    headers: new Headers({
      "Authorization": `Bearer ${token}`
    }),
  })
    .then(response => response.json())
    .then(data => {
      console.log(url);
      console.log(data);
      document.getElementById('inputName').value = data.name;
      document.getElementById('Email').value = data.email;

      const birthday = data.birthday.split('T')[0]; // Извлекаем только дату из значения birthday
      document.getElementById('inputData').value = birthday;
      document.getElementById('selectGender').value = data.gender;
      document.getElementById('inputTel').value = data.phone;
    })
    .catch(error => {
      console.error('Ошибка', error);
    });
}
const url = `https://mis-api.kreosoft.space/api/doctor/profile`;
get(url, token)

function populateSpecialties(specialties) {
  const selectSpecialties = document.getElementById('selectSpecialties');
  specialties.forEach(specialty => {
    const option = document.createElement('option');
    option.value = specialty.id;
    option.text = specialty.name;
    selectSpecialties.appendChild(option);
  });
}


async function changeProfilePut(data, token) {
  const url = 'https://mis-api.kreosoft.space/api/doctor/profile';
  return fetch(url, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      "Authorization": `Bearer ${token}`
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
    })
    .catch(error => {
      console.error('Ошибка', error);
      window.location.href = '../profile/profile.html';
    });
}

const form = document.querySelector('form');

const submitButton = document.getElementById('submitButton');
submitButton.addEventListener('click', function (event) {
  event.preventDefault();
  const inputFields = document.querySelectorAll('input, select');
  const isDisabled = inputFields[0].disabled;
  inputFields.forEach(function (field) {
    field.disabled = !isDisabled;
  });
  submitButton.textContent = isDisabled ? 'Сохранить изменения' : 'Изменить данные';

  if (!isDisabled) {
    const name = document.getElementById('inputName').value;
    const email = document.getElementById('Email').value;
    const birthday = document.getElementById('inputData').value;
    const gender = document.getElementById('selectGender').value;
    const phone = document.getElementById('inputTel').value;

    const data = {
      email: email,
      name: name,
      birthday: birthday,
      gender: gender,
      phone: phone
    };

    if (!phone) {
      data.phone = null;
    }

    console.log(data);

    changeProfilePut(data, token);
  }
});
