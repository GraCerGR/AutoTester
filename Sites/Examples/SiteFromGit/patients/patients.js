var token = localStorage.getItem('token');
console.log(token)
let page = 1;

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
      createCard(data);

      maxPagination = data.pagination.count;
      updatePagination(maxPagination);
    })
    .catch(error => {
      console.error('Ошибка', error);
    });
}
const url = `https://mis-api.kreosoft.space/api/patient`;
get(url, token)



function createCard(data) {
  const cardContainerWrapper = document.querySelector('.row.list');
  data.patients.forEach(patient => {

    const cardContainer = document.createElement('div');
    cardContainer.classList.add('col-lg-6');

    const container2 = document.createElement('div');
    container2.classList.add('container2', 'container', 'p-2', 'my-2');
    const queryString = new URLSearchParams(patient.id).toString();
    const url = `/patient/patient.html?${queryString}`;
    container2.innerHTML = `
          <h6>${patient.name}</h6>
          Пол: ${patient.gender === 'Male' ? 'Мужчина' : patient.gender === 'Female' ? 'Женщина' : 'Не указано'}<BR>
          Дата рождения - ${patient.birthday ? patient.birthday.split('T')[0] : 'Не указано'}<BR>
      `;

    cardContainer.appendChild(container2);
    cardContainerWrapper.appendChild(cardContainer);
    cardContainer.addEventListener('click', () => {
      window.location.href = url;
    });
  });
}


function searchPatients() {
  console.log(page);
  const name = document.getElementById('inputName').value;
  const conclusions = Array.from(document.getElementById('conclusionsSelect').selectedOptions).map(option => `conclusions=${option.value}`).join('&');
  const scheduledVisits = document.getElementById('scheduledVisitsSwitch').checked;
  const onlyMine = document.getElementById('myPatientsSwitch').checked;
  const sorting = document.getElementById('sortSelect').value;
  const pageSize = document.getElementById('pageSizeInput').value;

  const urlParams = new URLSearchParams();
  urlParams.set('name', name);
  urlParams.set('sorting', sorting);
  urlParams.set('scheduledVisits', scheduledVisits);
  urlParams.set('onlyMine', onlyMine);
  if (event && event.target.id === 'searchButton') {
    page = 1; // Установите значение page равным 1 при нажатии на кнопку searchButton
  }
  urlParams.set('page', page); // Используйте переданную страницу
  urlParams.set('size', pageSize);

  const url = `https://mis-api.kreosoft.space/api/patient?${conclusions}&${urlParams.toString()}`;

  const cardContainerWrapper = document.querySelector('.row.list');
  cardContainerWrapper.innerHTML = '';

  console.log(url);
  console.log(name);

  get(url, token);
}

function updatePagination(maxPagination) {
  const pagination = document.getElementById('pagination');
  pagination.innerHTML = '';
  const currentPage = parseInt(page); // текущая страница
  const totalPages = maxPagination; // общее количество страниц
  let startPage, endPage;
  if (totalPages <= 5) {
    // Если общее количество страниц меньше или равно 5, то отображаем все страницы
    startPage = 1;
    endPage = totalPages;
    console.log(page)
  } else {
    // Иначе вычисляем начальную и конечную страницы
    if (currentPage <= 3) {
      startPage = 1;
      endPage = 5;
      console.log(page)
    } else if (currentPage + 2 >= totalPages) {
      console.log(currentPage + 2)
      console.log(totalPages)
      startPage = totalPages - 4;
      endPage = totalPages;
      console.log(totalPages)
    } else {
      startPage = currentPage - 2;
      endPage = currentPage + 2;
      console.log(page)
    }
  }

  // Элементы для номеров страниц и добавление их в пагинацию
  for (let i = startPage; i <= endPage; i++) {
    const li = document.createElement('li');
    li.classList.add('page-item');
    const a = document.createElement('a');
    a.classList.add('page-link');
    a.href = '#';
    a.innerText = i;
    console.log(i);
    if (i === currentPage) {
      li.classList.add('active'); // выделяем текущую страницу
    }
    li.appendChild(a);
    pagination.appendChild(li);
  }

  const firstPageLink = document.createElement('a');
  firstPageLink.classList.add('page-link');
  firstPageLink.href = '#';
  firstPageLink.innerHTML = '&laquo;';
  const lastPageLink = document.createElement('a');
  lastPageLink.classList.add('page-link');
  lastPageLink.href = '#';
  lastPageLink.innerHTML = '&raquo;';

  const firstPageItem = document.createElement('li');
  firstPageItem.classList.add('page-item');
  firstPageItem.appendChild(firstPageLink);
  const lastPageItem = document.createElement('li');
  lastPageItem.classList.add('page-item');
  lastPageItem.appendChild(lastPageLink);

  pagination.insertBefore(firstPageItem, pagination.firstChild);
  pagination.appendChild(lastPageItem);
}


document.getElementById('searchButton').addEventListener('click', searchPatients);

pagination.addEventListener('click', (event) => {
  event.preventDefault();
  const link = event.target;
  if (!link.classList.contains('page-link')) {
    return;
  }
  console.log(maxPagination);
  if (link.innerText === '«') {
    page = page - 1; // Получаем предыдущий номер страницы
    console.log(page);
  } else if (link.innerText === '»') {
    console.log(page);
    page = parseInt(page) + 1; // Получаем следующий номер страницы
    console.log(page);
  }
  if (page <= 1) {
    page = 1;
  } else if (page >= maxPagination) {
    page = maxPagination;
    console.log(page);
  }
  if (link.innerText >= 1 && link.innerText <= maxPagination) {
    page = link.innerText; // Получите номер страницы из текста ссылки
  }
  console.log(page)

  searchPatients(page);
});


async function CreatePatientPost(data, token) {
  const url = 'https://mis-api.kreosoft.space/api/patient';
  return fetch(url, {
    method: 'POST',
    body: JSON.stringify(data),
    headers: new Headers({
      'Content-Type': 'application/json',
      "Authorization": `Bearer ${token}`
    })
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

      console.log(result);
    })
    .catch(error => {
      console.error('Ошибка', error);
      const errorMessage = document.getElementById('errorMessage');
      errorMessage.textContent = 'Произошла ошибка при создании пациента. Пожалуйста, попробуйте еще раз.';
    });
}


document.querySelector('#registrationModal .btn-primary').addEventListener('click', function () {
  const name = document.querySelector('#fullNameInput').value;
  const gender = document.querySelector('#genderSelect').value;
  const birthdate = document.querySelector('#birthdateInput').value;


  const patient = {
    name: name,
    birthday: birthdate + 'T00:00:00.000Z',
    gender: gender
  };

  CreatePatientPost(patient, token);

  const registrationModal = new bootstrap.Modal(document.getElementById('registrationModal'));
  registrationModal.hide();
});
