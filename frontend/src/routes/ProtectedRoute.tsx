import React from 'react';
import { Navigate } from 'react-router-dom';
import { getAccessToken } from '../auth';

export const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return getAccessToken() ? children : <Navigate to="/login" />;
};
