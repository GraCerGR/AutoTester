var token = localStorage.getItem('token');
let page = 1;

async function getDictionary(url) {
    return fetch(url, {
        method: 'GET',
    })
        .then(response => response.json())
        .then(data => {
            console.log(url);
            console.log(data);
            populateDictionary(data);
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}

async function getInspection(url, token) {
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
            createCard(data.inspections);

            maxPagination = data.pagination.count;
            updatePagination(maxPagination);
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}

function getChain(id, token) {
    console.log(id);
    const url = `https://mis-api.kreosoft.space/api/inspection/${id}/chain`;
    console.log(url);
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
            createCard(data, id);
        })
        .catch(error => {
            console.error('Ошибка', error);
        });

}


let queryString = window.location.search;
queryString = queryString.slice(1, -1);
const urlPatient = `https://mis-api.kreosoft.space/api/consultation?grouped=false&page=1&size=5`;
console.log(urlPatient);

const urlInspection = `https://mis-api.kreosoft.space/api/consultation?grouped=false&page=1&size=5`;
getInspection(urlInspection, token)

const urlDictionary = `https://mis-api.kreosoft.space/api/dictionary/icd10/roots`;
getDictionary(urlDictionary)


async function formatBirthday(originalDate) {
    const dateParts = originalDate.split("-");
    console.log(originalDate)
    const day = dateParts[2];
    const month = dateParts[1];
    const year = dateParts[0];
    originalDate = `${day}.${month}.${year}`;
    console.log(originalDate);
    return originalDate
}


function populateDictionary(dictionaries) {
    const selectDictionary = document.getElementById('selectDictionary');
    dictionaries.forEach(dictionary => {
        const option = document.createElement('option');
        option.value = dictionary.id;
        option.text = dictionary.name;
        selectDictionary.appendChild(option);
    });
}


function createCard(data, id) {
    let cardContainerWrapper;
    if (id) {
        cardContainerWrapper = document.getElementById(`collapse-${id}`);
    } else {
        cardContainerWrapper = document.querySelector('.row.list');
    }
    let marginLeft = 20;
    data.forEach(async inspection => {
        let arrow = '';
        let add = '';
        const cardContainer = document.createElement('div');
        const container2 = document.createElement('div');
        container2.classList.add('p-2', 'my-2');
        if (!id) {
            container2.classList.add('container2');
            cardContainer.classList.add('col-lg-6');
        } else {
            container2.classList.add('container3');
            cardContainer.classList.add('col-lg');
        }
        const conclusionText =
            inspection.conclusion === 'Recovery'
                ? 'Выздоровление'
                : inspection.conclusion === 'Disease'
                    ? 'Болезнь'
                    : inspection.conclusion === 'Death'
                        ? 'Смерть'
                        : 'Не указан';
        if (inspection.conclusion === 'Death')
            container2.classList.add('death');
        if (document.getElementById('grupSwitch').checked) {
            if (inspection.hasChain == true) {
                arrow = `<button type="button" class="btn btn-light" data-bs-toggle="collapse" data-bs-target="#collapse-${inspection.id}")">▼</button>`;
                getChain(inspection.id, token)
                console.log(inspection.id);
            }
            console.log(inspection.hasNested, inspection.hasChain);
            if ((inspection.hasNested == true && inspection.hasChain == false)) {
                arrow = `<button type="button" class="btn btn-light" data-bs-toggle="collapse" data-bs-target="#collapse-${inspection.id}")">▼</button>`;
                container2.classList.add('shifted');
                container2.style.marginLeft = `${marginLeft}px`;
                marginLeft += 20;
                if (marginLeft > 40) {
                    marginLeft = 40;
                }
            }

            if ((inspection.hasNested == false && inspection.hasChain == false && inspection.previousId)) {
                arrow = `<button type="button" class="btn btn-light disabled"">■</button>`;
                container2.classList.add('shifted');
                container2.style.marginLeft = `${marginLeft}px`;
                marginLeft += 20;
                if (marginLeft > 40) {
                    marginLeft = 40;
                }
            }
        }
        container2.innerHTML = `
      <div class="row">
      <div class="col-md-7">
        ${arrow}
        <div class="background">
          ${await formatBirthday(inspection.date.split('T')[0])}
        </div>
        <strong>Амбулаторный осмотр</strong>
      </div>
      <div class="col-md-5 d-flex justify-content-end align-items-center">
        <a href="../details/details.html?inspectionId=${inspection.id}&patient=${queryString}" class="btn add-inspection-link btn-outline-primary btn-sm">🔍 Детали осмотра</a>
      </div>
    </div>
    <div>Заключение: <strong>${conclusionText}</strong></div>
    <div>Основной диагноз: <strong>${inspection.diagnosis.name}</strong></div>
    <div class="fw-light">Медицинский работник: ${inspection.doctor}</div>
    <div id="collapse-${inspection.id}" class="collapse"></div>
    `;
        cardContainer.appendChild(container2);
        cardContainerWrapper.appendChild(cardContainer);
    });
}


function searchInspection() {
    console.log(page);
    const dictionary = Array.from(document.getElementById('selectDictionary').selectedOptions).map(option => `icdRoots=${option.value}`).join('&');
    const grupSwitch = document.getElementById('grupSwitch').checked;
    const pageSize = document.getElementById('pageSizeInput').value;

    const urlParams = new URLSearchParams();
    urlParams.set('grouped', grupSwitch);
    if (event && event.target.id === 'searchButton') {
        page = 1; // Установите значение page равным 1 при нажатии на кнопку searchButton
    }
    urlParams.set('page', page); // Используйте переданную страницу
    urlParams.set('size', pageSize);

    const url = `https://mis-api.kreosoft.space/api/consultation/?&${urlParams.toString()}`;

    const cardContainerWrapper = document.querySelector('.row.list');
    cardContainerWrapper.innerHTML = '';

    // Выполните нужные вам действия с полученной ссылкой
    console.log(url);
    console.log(name);

    getInspection(url, token);
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

    // Создаем элементы для номеров страниц и добавляем их в пагинацию
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


document.getElementById('searchButton').addEventListener('click', searchInspection);

pagination.addEventListener('click', (event) => {
    event.preventDefault();
    const link = event.target;
    if (!link.classList.contains('page-link')) {
        return;
    }
    console.log(maxPagination);
    if (link.innerText === '«') {
        //console.log(currentPage);
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

    searchInspection(page);
});