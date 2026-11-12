import { lazy, Suspense } from 'react';
import { Route, Routes } from 'react-router-dom';
import { PublicLayout } from '@/layouts/PublicLayout';
import { CabinetLayout } from '@/layouts/CabinetLayout';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { Loader } from '@/components/States';
import HomePage from '@/pages/HomePage';
import { LessonsPage, RentalPage } from '@/pages/CatalogPages';
import AuthPage from '@/pages/AuthPage';
import NotFoundPage from '@/pages/NotFoundPage';
import { BookLessonPage, BookRentalPage, CheckoutPage, PaymentResultPage } from '@/pages/BookingPages';

// Кабинеты грузим лениво — гостю они не нужны
const Client = lazy(() => import('@/pages/account/ClientPages').then((m) => ({ default: m.ClientHomePage })));
const ClientBookings = lazy(() => import('@/pages/account/ClientPages').then((m) => ({ default: m.ClientBookingsPage })));
const ClientProfile = lazy(() => import('@/pages/account/ClientPages').then((m) => ({ default: m.ClientProfilePage })));
const Instructor = lazy(() => import('@/pages/instructor/InstructorPages').then((m) => ({ default: m.InstructorHomePage })));
const InstructorSchedule = lazy(() => import('@/pages/instructor/InstructorPages').then((m) => ({ default: m.InstructorSchedulePage })));
const InstructorProfile = lazy(() => import('@/pages/instructor/InstructorPages').then((m) => ({ default: m.InstructorProfilePage })));
const Admin = lazy(() => import('@/pages/admin/AdminPages').then((m) => ({ default: m.AdminHomePage })));
const AdminUsers = lazy(() => import('@/pages/admin/AdminPages').then((m) => ({ default: m.AdminUsersPage })));
const AdminCatalog = lazy(() => import('@/pages/admin/AdminPages').then((m) => ({ default: m.AdminCatalogPage })));
const AdminBookings = lazy(() => import('@/pages/admin/AdminPages').then((m) => ({ default: m.AdminBookingsPage })));

export default function App() {
  return (
    <Suspense fallback={<div className="container section"><Loader /></div>}>
      <Routes>
        <Route element={<PublicLayout />}>
          <Route index element={<HomePage />} />
          <Route path="lessons" element={<LessonsPage />} />
          <Route path="rental" element={<RentalPage />} />
          <Route path="book/lesson/:id?" element={<BookLessonPage />} />
          <Route path="book/rental/:id?" element={<BookRentalPage />} />
          <Route path="login" element={<AuthPage mode="login" />} />
          <Route path="register" element={<AuthPage mode="register" />} />
          <Route element={<ProtectedRoute roles={['client']} />}>
            <Route path="checkout/:id" element={<CheckoutPage />} />
            <Route path="checkout/:id/result" element={<PaymentResultPage />} />
          </Route>
          <Route path="*" element={<NotFoundPage />} />
        </Route>

        <Route element={<ProtectedRoute roles={['client']} />}>
          <Route path="account" element={<CabinetLayout />}>
            <Route index element={<Client />} />
            <Route path="bookings" element={<ClientBookings />} />
            <Route path="profile" element={<ClientProfile />} />
          </Route>
        </Route>

        <Route element={<ProtectedRoute roles={['instructor']} />}>
          <Route path="instructor" element={<CabinetLayout />}>
            <Route index element={<Instructor />} />
            <Route path="schedule" element={<InstructorSchedule />} />
            <Route path="profile" element={<InstructorProfile />} />
          </Route>
        </Route>

        <Route element={<ProtectedRoute roles={['admin']} />}>
          <Route path="admin" element={<CabinetLayout />}>
            <Route index element={<Admin />} />
            <Route path="users" element={<AdminUsers />} />
            <Route path="catalog" element={<AdminCatalog />} />
            <Route path="bookings" element={<AdminBookings />} />
          </Route>
        </Route>
      </Routes>
    </Suspense>
  );
}
