# Meclis Oylama ve Biyometrik Yoklama Sistemi

Bu proje; meclis, kurul veya komite gibi karar organlarının toplantılarını dijitalleştirmek, gerçek zamanlı (real-time) oylama süreçlerini yönetmek ve biyometrik yüz tanıma simülasyonu ile salondaki üye varlığını otomatik takip etmek amacıyla geliştirilmiş çok katmanlı (multi-tier) bir sistemdir.

## 🚀 Teknolojik Altyapı

* **Backend:** Go 1.25, Gin Web Framework, WebSockets (`golang.org/x/net/websocket`), PostgreSQL
* **Bridge (Köprü):** Python >= 3.12, `uv` paket yöneticisi, Standart urllib
* **Frontend:** Pure HTML5, CSS3 (Modern CSS Değişkenleri), Vanilla JavaScript

---

## 📂 Proje Yapısı

Sistem üç ana bileşenden oluşur:
1.  **`backend/`**: API uç noktalarını sunan, WebSocket Hub mimarisiyle oyları anlık dağıtan ve veritabanı ilişkilerini yöneten Go servisidir.
2.  **`bridge/`**: Biyometrik donanım entegrasyon katmanıdır. Donanım henüz bağlı değilken salona giren/çıkan üyeleri otomatik simüle ederek ana sunucuya HTTP üzerinden raporlar.
3.  **`frontend/`**: Moderatör (Meclis Başkanı) ve Üyeler için özelleştirilmiş, WebSocket bağlantısıyla sayfayı yenilemeden çalışan canlı arayüzlerdir.

---

## 🛠️ Kurulum ve Çalıştırma

Sistemi yerel ortamınızda ayağa kaldırmak için aşağıdaki adımları sırasıyla uygulayın.

### 1. Veritabanının Başlatılması (Docker)
Lokalinizdeki diğer PostgreSQL servisleriyle çakışma yaşanmaması adına veritabanı **5433** portundan dışarı açılmaktadır.

* **Windows (PowerShell) için tek satır komut:**
    ```powershell
    docker run --name meclis-db -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=meclis -p 5433:5432 -d postgres:15-alpine
    ```
* **Linux / macOS (Terminal) veya çok satırlı PowerShell için:**
    ```bash
    docker run --name meclis-db `
      -e POSTGRES_USER=postgres `
      -e POSTGRES_PASSWORD=postgres `
      -e POSTGRES_DB=meclis `
      -p 5433:5432 `
      -d postgres:15-alpine
    ```

### 2. Backend Servisinin Başlatılması (Go)
1.  `backend/` dizinine geçiş yapın:
    ```bash
    cd backend
    ```
2.  Uygulamanın 5433 portundaki PostgreSQL'e bağlanabilmesi için `backend/` klasörünün altında `.env` isimli bir dosya oluşturup şu satırı ekleyin:
    ```env
    DATABASE_URL=postgres://postgres:postgres@localhost:5433/meclis?sslmode=disable
    ```
3.  Veritabanı tablolarını oluşturmak ve test verilerini (seed) yüklemek için migration script'ini çalıştırın:
    ```bash
    go run migrate_run.go
    ```
4.  Ana backend sunucusunu başlatın (Varsayılan olarak `8080` portunda ayağa kalkar):
    ```bash
    go run main.go
    ```

### 3. Yüz Tanıma Köprüsünün Başlatılması (Python)
Sistemin çalışabilmesi için salonda aktif üyelerin bulunması gerekir. `bridge` servisi bunu otomatik simüle eder.
1.  `bridge/` dizinine geçiş yapın:
    ```bash
    cd bridge
    ```
2.  Modern Python paket yöneticisi `uv` kullanarak simülatörü çalıştırın:
    ```bash
    uv run bridge
    ```
    *Bu servis çalıştığı sürece konsolda üyelerin giriş/çıkış yaptığını (`✓ üye 3 → entry`) göreceksiniz.*

### 4. Frontend Arayüzünün Başlatılması
CORS güvenlik politikaları gereği arayüzün tam olarak `http://localhost:3000` adresinde çalışması şarttır.
1.  `frontend/` dizinine geçin ve statik bir HTTP sunucu başlatın:
    ```bash
    cd frontend
    # Node.js yüklüyse hızlıca başlatmak için:
    npx serve -p 3000
    ```

---

## 🕹️ Adım Adım Lokal Test Senaryosu

Sistemin tüm parçalarını başarıyla çalıştırdıktan sonra uçtan uca test etmek için şu adımları takip edin:

1.  Tarayıcınızdan `http://localhost:3000` adresine gidin.
2.  **Moderatör Girişi:** TC Kimlik alanına `10000000000` yazıp herhangi bir şifre girerek giriş yapın. 
    * Moderatör panelindeki "Salondaki Üye" sayısı Python simülatörü nedeniyle dinamik olarak artacaktır.
    * Toplantı yeter sayısı olan **16 üye** salona giriş yaptığında "Oylama Başlat" butonu aktif hale gelecektir.
3.  **Üye Girişi:** Tarayıcınızdan gizli bir sekme (veya farklı bir tarayıcı) açıp tekrar `http://localhost:3000` adresine gidin. TC Kimlik alanına `10000000001` yazarak sisteme normal üye olarak giriş yapın.
    * *(Not: Eğer simülatör o esnada ilgili üyeyi henüz salona sokmadıysa "salonda değilsiniz" uyarısı alabilirsiniz. 1-2 saniye bekleyip tekrar deneyin).*
4.  **Oylama Süreci:** * Moderatör ekranından bir gündem maddesi yazıp (Örn: "2026 Yılı Bütçe Planı") **Oylama Başlat** butonuna basın.
    * Üyenin açık olduğu diğer ekranda, hiçbir sayfayı yenilemeden otomatik olarak **EVET / HAYIR** butonları belirecektir.
    * Üye ekranından oy kullanıldığında, moderatör ekranındaki oy sayacı WebSocket Hub üzerinden anlık olarak güncellenecektir.
5.  **Kapatma:** Moderatör panelinden **Oylamayı Kapat** butonuna basılarak oylama mühürlenir ve karar sonucu tüm ekranlara eşzamanlı yansıtılır.