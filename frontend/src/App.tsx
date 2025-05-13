import React from 'react';
import { Routes, Route } from 'react-router-dom';
import Navbar from './components/Navbar';
import Home from './pages/Home';
import Essays from './pages/Essays';
import EssayDetail from './pages/EssayDetail';
import Reviews from './pages/Reviews';
import Login from './pages/Login';
import Register from './pages/Register';
import MyEssay from './pages/MyEssay';
import MyReviews from './pages/MyReviews';
import { ProtectedRoute } from './routes/ProtectedRoute';
import { Container, CssBaseline } from '@mui/material';

const App: React.FC = () => (
  <>
    <CssBaseline />
    <Navbar />
    <Container sx={{ mt: 4 }}>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/essays" element={<Essays />} />
        <Route path="/essay/:author" element={<EssayDetail />} />
        <Route path="/reviews" element={<Reviews />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/my-essay" element={<ProtectedRoute><MyEssay /></ProtectedRoute>} />
        <Route path="/my-reviews" element={<ProtectedRoute><MyReviews /></ProtectedRoute>} />
      </Routes>
    </Container>
  </>
);

export default App;
