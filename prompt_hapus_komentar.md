Berikut adalah prompt yang bisa kamu salin dan tempelkan (copy-paste) ke AI untuk mengerjakan fitur hapus komentar di Backend dan panduan untuk Frontend-nya:

---

**Prompt untuk Backend:**

> "Tolong tambahkan fitur agar Admin bisa menghapus komentar pada forum (untuk moderasi/hate speech). Lakukan perubahan berikut:
> 
> 1. Buka `controllers/comment.go`, lalu tambahkan method `DeleteComment` pada `CommentController`. Method ini harus:
>    - Mengambil parameter `id` komentar dari URL (`mux.Vars(r)`).
>    - Melakukan query `DELETE FROM comments WHERE id = $1`.
>    - Mengembalikan response error 404 jika komentar tidak ditemukan, atau error 500 jika gagal eksekusi query.
>    - Mengembalikan response JSON `{"message": "Komentar berhasil dihapus"}` dengan HTTP Status 200 OK jika berhasil.
> 
> 2. Buka `routes/routes.go`, di bagian inisiasi routes admin (yang menggunakan prefix `/admin` dan dilindungi oleh auth admin middleware), tambahkan endpoint untuk menghapus komentar:
>    `admin.HandleFunc("/comments/{id:[0-9]+}", commentController.DeleteComment).Methods("DELETE")`
>
> Tolong berikan saya kode yang telah diperbarui untuk file-file tersebut."

---

**Panduan untuk Frontend (Admin UI):**

Jika kamu ingin meminta AI untuk memodifikasi Frontend (karena repo Frontend sepertinya terpisah dari sini), gunakan prompt berikut:

> "Tolong tambahkan tombol 'Hapus' di setiap daftar komentar pada halaman dashboard admin/forum management. Ketika tombol hapus ditekan:
> 1. Tampilkan konfirmasi `window.confirm('Apakah Anda yakin ingin menghapus komentar ini?')`.
> 2. Jika disetujui, panggil API `DELETE /api/admin/comments/{id}` (jangan lupa sertakan kredensial auth / JWT cookies).
> 3. Jika berhasil, tampilkan notifikasi sukses dan muat ulang (refetch) daftar komentar forum tersebut."
