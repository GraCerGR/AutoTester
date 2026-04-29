import os
from requests import options
from selenium import webdriver
from selenium.webdriver.chrome.options import Options as ChromeOptions

# Переопределяем TEST_URL, если передано через окружение
TEST_URL = os.environ.get("TEST_URL", "http://localhost:8080")
__all__ = ["TEST_URL"]

# Берём адрес Selenium Hub из переменных окружения
SELENIUM_HUB = os.environ.get("SELENIUM_HUB")
SESSION_NAME = os.environ.get("SESSION_NAME")

if SELENIUM_HUB:
    # Сохраняем оригинальный класс, если понадобится
    OriginalChrome = webdriver.Chrome

    class PatchedChrome(webdriver.Remote):
        def __init__(self, *args, **kwargs):
            # Опции Chrome
            options = kwargs.pop("options", ChromeOptions())
            options.add_argument("--unsafely-treat-insecure-origin-as-secure=http://worker1,http://worker2,http://worker3,http://worker4,http://worker5")
            options.add_argument("--allow-insecure-localhost")
            #options.add_argument("--headless=new")  # Без GUI
            #options.add_argument("--no-sandbox")
            #options.add_argument("--disable-dev-shm-usage")
            #options.add_argument("--window-size=1280,800")
            
            if SESSION_NAME:
                options.set_capability("se:name", SESSION_NAME)
                  
            # Передаём в Remote
            super().__init__(command_executor=SELENIUM_HUB, options=options, *args, **kwargs)

    # Переназначаем webdriver.Chrome на патч
    webdriver.Chrome = PatchedChrome