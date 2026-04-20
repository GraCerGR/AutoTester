package org.openqa.selenium.chrome;

import java.net.MalformedURLException;
import java.net.URL;

import org.openqa.selenium.Capabilities;
import org.openqa.selenium.remote.RemoteWebDriver;

public class ChromeDriver extends RemoteWebDriver {

    public static final String TEST_URL =
        getEnvOrDefault("TEST_URL", "http://localhost:8080");

    public static final String SELENIUM_HUB =
        getEnv("SELENIUM_HUB");

    public static final String SESSION_NAME =
        getEnv("SESSION_NAME");

    public ChromeDriver() {
        this(buildOptionsWithEnv());
    }

    public ChromeDriver(Capabilities capabilities) {
        super(getHubURL(), mergeCapabilities(capabilities));
    }

    public ChromeDriver(ChromeOptions options) {
        super(getHubURL(), enrichOptions(options));
    }

    private static ChromeOptions buildOptionsWithEnv() {
        ChromeOptions options = new ChromeOptions();
        return enrichOptions(options);
    }

    private static ChromeOptions enrichOptions(ChromeOptions options) {
        if (options == null) {
            options = new ChromeOptions();
        }

        // Аналогично sitecustomize.py
        // options.addArguments("--headless=new");
        // options.addArguments("--no-sandbox");
        // options.addArguments("--disable-dev-shm-usage");
        options.addArguments("--window-size=1280,800");

        if (SESSION_NAME != null && !SESSION_NAME.trim().isEmpty()) {
            options.setCapability("se:name", SESSION_NAME.trim());
        }

        return options;
    }

    private static Capabilities mergeCapabilities(Capabilities capabilities) {
        ChromeOptions options = new ChromeOptions();

        if (capabilities != null) {
            options.merge(capabilities);
        }

        return enrichOptions(options);
    }

    private static URL getHubURL() {
        if (SELENIUM_HUB == null || SELENIUM_HUB.trim().isEmpty()) {
            throw new RuntimeException(
                "SELENIUM_HUB is not set. Set SELENIUM_HUB to the Selenium Grid URL, e.g. http://selenium-hub:4444"
            );
        }

        String hub = SELENIUM_HUB.trim();

        // Нормализуем URL: если нет /wd/hub — добавляем
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

    private static String getEnv(String name) {
        String value = System.getenv(name);
        if (value == null) {
            return null;
        }
        value = value.trim();
        return value.isEmpty() ? null : value;
    }

    private static String getEnvOrDefault(String name, String defaultValue) {
        String value = getEnv(name);
        return value == null ? defaultValue : value;
    }
}