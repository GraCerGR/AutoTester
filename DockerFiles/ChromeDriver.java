// src/test/java/org/selenium/chrome/ChromeDriver.java
package org.openqa.selenium.chrome;

import java.net.MalformedURLException;
import java.net.URL;

import org.openqa.selenium.Capabilities;
import org.openqa.selenium.chrome.ChromeOptions;
import org.openqa.selenium.remote.RemoteWebDriver;

/**
 * Shim: класс с тем же FQN, что и настоящий ChromeDriver, но
 * создаёт RemoteWebDriver, подключённый к Selenium Hub.
 *
 * Требования:
 *  - В проекте должна быть зависимость selenium-java (чтобы были RemoteWebDriver, ChromeOptions и т.д.)
 *  - В контейнере должна быть доступна переменная окружения SELENIUM_HUB, например
 *      SELENIUM_HUB=http://selenium-hub:4444
 *
 * Поведение:
 *  - new ChromeDriver() и new ChromeDriver(ChromeOptions) будут открывать сессию на Hub.
 *  - Если SELENIUM_HUB не задана — падаем с RuntimeException (чтобы явно видно было ошибку), или можно поменять на fallback.
 */
public class ChromeDriver extends RemoteWebDriver {

    public ChromeDriver() {
        this(new ChromeOptions());
    }

    public ChromeDriver(Capabilities capabilities) {
        super(getHubURL(), capabilities);
    }

    public ChromeDriver(ChromeOptions options) {
        super(getHubURL(), options);
    }

    private static URL getHubURL() {
        String hub = System.getenv("SELENIUM_HUB");
        if (hub == null || hub.trim().isEmpty()) {
            throw new RuntimeException("SELENIUM_HUB is not set. Set SELENIUM_HUB to the Selenium Grid URL, e.g. http://selenium-hub:4444");
        }

        hub = hub.trim();

        // Нормализуем: если пользователь указал базовый URL без /wd/hub — добавим его.
        // Если уже указан путь — оставим как есть.
        if (!hub.endsWith("/wd/hub") && !hub.contains("/wd/hub")) {
            if (hub.endsWith("/")) {
                hub = hub + "wd/hub";
            } else {
                hub = hub + "/wd/hub";
            }
        }

        try {
            return new URL(hub);
        } catch (MalformedURLException e) {
            throw new RuntimeException("Invalid SELENIUM_HUB URL: " + hub, e);
        }
    }
}
