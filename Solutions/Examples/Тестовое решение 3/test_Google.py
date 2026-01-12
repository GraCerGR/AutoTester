from contextlib import nullcontext

import pytest
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

SEARCH_QUERY = "OpenAI ChatGPT"
GOOGLE_URL = "https://code.hits.university/"

def find_first_button(driver, query: str, timeout: int = 10):
    driver.get(GOOGLE_URL)

    # Ожидание загрузки страницы
    TSUButton = WebDriverWait(driver, 10).until(
        EC.element_to_be_clickable((By.XPATH, "/html/body/div/div/div[1]/form/div/div[3]/a"))
    )

    # Ввод данных ТГУ
    EnterButton = WebDriverWait(driver, 10).until(
        EC.element_to_be_clickable((By.XPATH, "/html/body/div/div/div[1]/form/div/div[3]/button"))
    )


    return EnterButton

def find_second_button(driver, query: str, timeout: int = 10):
    driver.get(GOOGLE_URL)

    TSUButton = WebDriverWait(driver, 10).until(
        EC.element_to_be_clickable((By.XPATH, "/html/body/div/div/div[1]/form/div/div[3]/a"))
    )
    TSUButton.click()

    # Ввод данных ТГУ
    EnterButton = WebDriverWait(driver, 10).until(
        EC.element_to_be_clickable((By.XPATH, "/html/body/div/main/div/div/div/div/section/form/div[3]/input[4]"))
    )

    return EnterButton

# Фикстура: создаём браузер один раз для модуля тестов
@pytest.fixture(scope="module")
def driver():
    chrome_opts = Options()
    # Если хотите запускать в фоне — раскомментируйте следующую строку:
    # chrome_opts.add_argument("--headless=new")
    chrome_opts.add_argument("--window-size=1280,800")
    #service = Service(ChromeDriverManager().install())
    #driver = webdriver.Chrome(service=service, options=chrome_opts)
    driver = webdriver.Chrome(options=chrome_opts)
    try:
        yield driver
    finally:
        driver.quit()

def test_first_button(driver):
    results = find_first_button(driver, SEARCH_QUERY)
    print("Найдено кнопок: ", results)
    assert results != None, "Ожидалась первая кнопка входа, но её нет"

def test_second_button(driver):
    results = find_second_button(driver, SEARCH_QUERY)
    print("Найдено кнопок: ", results)
    assert results != None, "Ожидалась вторая кнопка входа, но её нет"

    # Получаем ссылку из первого результата
    first_anchor = results[0].find_element(By.TAG_NAME, "a")
    href = first_anchor.get_attribute("href") or ""
    # Для устойчивости приведём в нижний регистр
    assert "openai.com" in href.lower(), f"Первая ссылка не ведёт на openai.com: {href}"
