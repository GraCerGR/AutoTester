import { createCard } from "../patient.js";


//Неполное
let incomplete_testing_data = [
{
    conclusion: "Death",
    date: "2024-11-06T08:16:00",
    diagnosis: 
    {  name: "Остеопороз с патологическим переломом после удаления яичников" },
    doctor: "Доктор Анестезиолог-реаниматологг",
    hasChain: false,
    hasNested: false,
    id: "6911d81c-30d7-4c5b-ba72-08dcfe20c9d6",
    previousId: null,
},
{
    conclusion: "Disease",
    date: "2024-06-30T05:05:00",
    diagnosis: 
    {  name: "Аборт неуточненный неполный, осложнившийся длительным или чрезмерным кровотечением" },
    doctor: "hendo",
    hasChain: true,
    hasNested: true,
    id: "3109a33d-8956-41d0-7161-08dc9a4cb881",
    previousId: null,
}
]

//Разделение на классы эквивалентновсти
// conclusion, Группировка, hasChain, hasNested, previousId
let division_into_equivalence_classes = [
{
    conclusion: "Death",
    date: "2025-11-06T08:16:00",
    diagnosis: 
    {  name: "Остеопороз" },
    doctor: "Доктор Дре",
    hasChain: false,
    hasNested: false,
    id: "6911d81c-30d7-4c5b-ba72-08dcfe20c9d6",
    previousId: null,
},
{
    conclusion: "Disease",
    date: "2024-06-30T05:05:00",
    diagnosis: 
    {  name: "Абортище" },
    doctor: "hendo",
    hasChain: true,
    hasNested: true,
    id: "3109a33d-8956-41d0-7161-08dc9a4cb881",
    previousId: null,
},
{
    conclusion: "Recovery",
    date: "2024-08-01T10:08:00",
    diagnosis: 
    {  name: "Денге" },
    doctor: "Доктор Анестезиолог",
    hasChain: false,
    hasNested: true,
    id: "0f7450e0-841c-40cf-7162-08dc9a4cb881",
    previousId: "3109a33d-8956-41d0-7161-08dc9a4cb881",
},
{
    conclusion: "Recovery",
    date: "2024-08-01T10:08:00",
    diagnosis: 
    {  name: "Денге" },
    doctor: "Доктор Анестезиолог",
    hasChain: true,
    hasNested: false,
    id: "213949be-8ae9-49e6-79e0-08dc99c44493",
    previousId: null,
},
{
    conclusion: "Disease",
    date: "2025-11-06T08:16:00",
    diagnosis: 
    {  name: "Остеопороз" },
    doctor: "Доктор Дре",
    hasChain: false,
    hasNested: false,
    id: "6911d81c-30d7-4c5b-ba72-08dcfe20c9d6",
    previousId: "3109a33d-8956-41d0-7161-08dc9a4cb881",
},
]

// Угадывание ошибок
let guessing_errors_data = [
    {
        conclusion: '', // Пустая строка
        date: '2024-01-01T00:00:00',
        diagnosis: { name: 'Заболевание 1' },
        doctor: '', // Пустая строка
        hasChain: false,
        hasNested: false,
        id: '1',
        previousId: '', // Пустая строка
      },
      {
        conclusion: 'NoExist', // Несуществующее заключение
        date: '2024-01-02T00:00:00',
        diagnosis: { name: 'Заболевание 2' },
        doctor: 'Доктор 2',
        hasChain: false,
        hasNested: false,
        id: '2',
        previousId: 'invalid-id', // Неверный формат ID
      },
    ]

//Классы плохих данных
let сlasses_of_bad_data_data = [
    [{
        conclusion: null, // Отсутствие заключения
        date: null, // Отсутствие даты
        diagnosis: { name: null }, // Отсутствие диагноза
        doctor: null, // Отсутствие врача
        hasChain: false,
        hasNested: false,
        id: null, // Отсутствие ID
        previousId: null,
      }],
      [{
        conclusion: 'Recovery',
        date: '', // Неверный тип
        diagnosis: { name: 'Заболевание 1' },
        doctor: 'Доктор 1',
        hasChain: 'yes', // Неверный тип
        hasNested: false,
        id: 123, // Неверный тип
        previousId: null,
      }],
      [{ // Неверный тип
        1: 'Recovery',
        2: '2024-01-01T00:00:00',
        3: { name: 'Заболевание 1' },
        4: 'Доктор 1',
        5: 'yes',
        6: false,
        7: 123,
        8: null,
      }],
    ]

const emptyData = [];
const noInicializationData = undefined

//Классы хороших данных
let сlasses_of_good_data_data = [
    [// Номинальные случаи
    {
      conclusion: 'Recovery',
      date: '2024-01-01T00:00:00',
      diagnosis: { name: 'Заболевание 1' },
      doctor: 'Доктор 1',
      hasChain: false,
      hasNested: false,
      id: '1',
      previousId: null,
    },
    {
      conclusion: 'Disease',
      date: '2024-01-02T00:00:00',
      diagnosis: { name: 'Заболевание 2' },
      doctor: 'Доктор 2',
      hasChain: true,
      hasNested: false,
      id: '2',
      previousId: null,
    },
    {
      conclusion: 'Death',
      date: '2024-01-03T00:00:00',
      diagnosis: { name: 'Заболевание 3' },
      doctor: 'Доктор 3',
      hasChain: false,
      hasNested: true,
      id: '3',
      previousId: null,
    },
  ],
  [ // Минимальная нормальная конфигурация
    {
      conclusion: 'Recovery',
      date: '2024-01-01T00:00:00',
      diagnosis: { name: 'Заболевание 1' },
      doctor: 'Доктор 1',
      hasChain: false,
      hasNested: false,
      id: '1',
      previousId: null,
    },
  ],
  [ // Максимальная нормальная конфигурация
    {
      conclusion: 'Recovery',
      date: '2024-01-01T00:00:00',
      diagnosis: { name: 'Заболевание 1' },
      doctor: 'Доктор 1',
      hasChain: true,
      hasNested: true,
      id: '1',
      previousId: '0',
    },
    {
      conclusion: 'Disease',
      date: '2024-01-02T00:00:00',
      diagnosis: { name: 'Заболевание 2' },
      doctor: 'Доктор 2',
      hasChain: true,
      hasNested: true,
      id: '2',
      previousId: '1',
    },
    {
      conclusion: 'Death',
      date: '2024-01-03T00:00:00',
      diagnosis: { name: 'Заболевание 3' },
      doctor: 'Доктор 3',
      hasChain: true,
      hasNested: true,
      id: '3',
      previousId: '2',
    },
    {
      conclusion: 'Recovery',
      date: '2024-01-04T00:00:00',
      diagnosis: { name: 'Заболевание 4' },
      doctor: 'Доктор 4',
      hasChain: true,
      hasNested: true,
      id: '4',
      previousId: '3',
    },
    {
      conclusion: 'Disease',
      date: '2024-01-05T00:00:00',
      diagnosis: { name: 'Заболевание 5' },
      doctor: 'Доктор 5',
      hasChain: true,
      hasNested: true,
      id: '5',
      previousId: '4',
    },
  ]];

let testing_on_data_streams_data =  [
    { //Инициализация
        conclusion: 'Recovery',
        date: '2024-01-01T00:00:00',
        diagnosis: { name: 'Заболевание 1' },
        doctor: 'Доктор 1',
        hasChain: false,
        hasNested: false,
        id: '1',
        previousId: null,
      },
      {//Использование
        conclusion: 'Disease',
        date: '2024-01-02T00:00:00',
        diagnosis: { name: 'Заболевание 2' },
        doctor: 'Доктор 2',
        hasChain: true,
        hasNested: false,
        id: '2',
        previousId: null,
      },
      {
        conclusion: 'Death',
        date: '2024-01-03T00:00:00',
        diagnosis: { name: 'Заболевание 3' },
        doctor: 'Доктор 3',
        hasChain: false,
        hasNested: true,
        id: '3',
        previousId: '2',
      },
      {//Доступ к переменным (переменные используются неправильно)
        conclusion: 'Recovery',
        date: null, // Дата не определена
        diagnosis: { name: 'Заболевание 1' },
        doctor: 'Доктор 1',
        hasChain: false,
        hasNested: false,
        id: '1',
        previousId: null,
      },
    ]

//Структурированное базисное (Оно не будет отличаться от Разделение на классы эквивалентновсти) - Нет
//Граничные значения (У метода нет граничных значений) - Нет

function Checker(grupSwitch, allShowSwitch){
document.getElementById('grupSwitch').checked = grupSwitch;
document.getElementById('allShowSwitch').checked = allShowSwitch;
}

async function Clean(){
const cardContainerWrapper = document.querySelector('.row.list');
while (cardContainerWrapper.firstChild) {
    cardContainerWrapper.removeChild(cardContainerWrapper.firstChild);
}
}


document.getElementById('btn1').addEventListener('click', function() {
    Clean()
    Checker(false, true);
    createCard(incomplete_testing_data);
});

document.getElementById('btn2').addEventListener('click', function() {
    Clean()
    Checker(false, true);
    createCard(testing_on_data_streams_data);
});

document.getElementById('btn3').addEventListener('click', function() {
    Clean()
    Checker(false, true);
    createCard(division_into_equivalence_classes);
});

document.getElementById('btn4').addEventListener('click', function() {
    Clean()
    Checker(true, false);
    createCard(division_into_equivalence_classes);
});

document.getElementById('btn5').addEventListener('click', function() {
    Clean()
    Checker(false, true);
    createCard(guessing_errors_data);
});

document.getElementById('btn6').addEventListener('click', function() {
    Clean()
    Checker(false, true);
    createCard(сlasses_of_bad_data_data[0]);
    createCard(сlasses_of_bad_data_data[1]);
    createCard(сlasses_of_bad_data_data[2]);
    createCard(emptyData);
    createCard(noInicializationData);
});

document.getElementById('btn7').addEventListener('click', function() {
    Clean()
    Checker(false, true);
    createCard(сlasses_of_good_data_data[0]);
    createCard(сlasses_of_good_data_data[1]);
    createCard(сlasses_of_good_data_data[2]);
});