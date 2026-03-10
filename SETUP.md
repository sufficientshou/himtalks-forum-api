# Setup Himtalks Backend

Panduan singkat untuk menjalankan project setelah clone dari GitHub.

## 1. File Environment (.env)

File `.env` sudah disertakan dengan nilai default. **Edit file `.env`** dan isi nilai yang sesuai:

### Wajib diisi untuk development:
- `SECRET_KEY` - Buat string random minimal 32 karakter (untuk JWT)
- `GOOGLE_CLIENT_ID` & `GOOGLE_CLIENT_SECRET` - Dari [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
- `GOOGLE_REDIRECT_URL` - Untuk local: `http://localhost:8080/auth/google/callback`

### Opsional (untuk fitur Songfess):
- `SPOTIFY_CLIENT_ID` & `SPOTIFY_CLIENT_SECRET` - Dari [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)

### Database:
- Default: `postgres` / `postgres` / `himtalks` di `localhost:5432`
- Sesuaikan jika PostgreSQL Anda berbeda

## 2. Menjalankan Aplikasi

### Opsi A: Dengan Docker (Recommended)

```bash
docker-compose up -d
```

PostgreSQL dan aplikasi akan berjalan. Akses API di `http://localhost:8080`

### Opsi B: Tanpa Docker (PostgreSQL harus sudah terinstall)

1. Pastikan PostgreSQL berjalan di `localhost:5432`
2. Buat database: `CREATE DATABASE himtalks;`
3. Jalankan aplikasi:

```bash
go run main.go
```

## 3. Endpoint Utama

- **Health**: `GET http://localhost:8080/`
- **Google Login**: `GET http://localhost:8080/auth/google/login`
- **Admin Setup**: Lihat `MANUAL_ADMIN_SETUP.md` untuk menambah admin pertama
