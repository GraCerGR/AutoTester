package org.example;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.Test;
import org.openqa.selenium.*;
import org.openqa.selenium.chrome.ChromeDriver;
import org.openqa.selenium.chrome.ChromeOptions;

import static org.junit.jupiter.api.Assertions.*;

public class PizzaSiteTest {

    private static WebDriver driver;

    // URL сайта
    private static final String TEST_URL = "http://localhost:5500";
    // поменяй при необходимости



    @BeforeAll
    static void setup() {

        ChromeOptions options = new ChromeOptions();

        // Можно включить headless
        // options.addArguments("--headless=new");

        options.addArguments("--window-size=1280,800");

        driver = new ChromeDriver(options);
    }


    @AfterAll
    static void teardown() {
        if (driver != null) {
            driver.quit();
        }
    }


    @Test
    void testPageHasTitle() {

        driver.get(TEST_URL);

        String title = driver.getTitle();

        assertFalse(title.isEmpty(), "У страницы отсутствует title");

        System.out.println("✅ Title найден: " + title);
    }


    @Test
    void testHeaderExists() {

        driver.get(TEST_URL);

        // В header у нас текст "🍕 Pizza Shop"
        WebElement header = driver.findElement(By.tagName("header"));

        String expectedText = "Pizza Shop";

        assertTrue(
                header.getText().contains(expectedText),
                "Ожидался текст '" + expectedText + "', но найден '" + header.getText() + "'"
        );

        System.out.println("✅ Заголовок найден: " + header.getText());
    }


    @Test
    void testPaginationWorks() {

        driver.get(TEST_URL);

        // Кнопка страницы 2
        WebElement page2Btn = driver.findElement(
                By.xpath("//button[text()='2']")
        );

        page2Btn.click();

        // Небольшая пауза для обновления DOM
        try {
            Thread.sleep(500);
        } catch (InterruptedException ignored) {}


        // Проверяем, что кнопка 2 стала активной
        WebElement activeBtn = driver.findElement(
                By.cssSelector(".pagination button.active")
        );

        assertEquals("2", activeBtn.getText(),
                "Страница 2 не стала активной");

        System.out.println("✅ Пагинация работает: открыта страница 2");
    }
}