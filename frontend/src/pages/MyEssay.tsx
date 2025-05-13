import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../api';
import {
  Box,
  Typography,
  TextField,
  Button,
  Alert,
  CircularProgress,
} from '@mui/material';

interface EssayResponse {
  id: number;
  content: string;
  author: string;
  created_at: string;
}

interface EssayRequest {
  content: string;
  author: string;
}

const fetchMyEssay = (username: string) =>
  api.get<EssayResponse>(`/essay/${username}`).then(r => r.data);

const publishEssay = (req: EssayRequest) =>
  api.post<EssayResponse>('/essay', req).then(r => r.data);

const MyEssay: React.FC = () => {
  const username = localStorage.getItem('username')!;
  const queryClient = useQueryClient();

  // Fetch existing essay (if any)
  const { data, isLoading } = useQuery<EssayResponse, Error>({
    queryKey: ['myEssay', username],
    queryFn: () => fetchMyEssay(username),
  });

  // Local state for the form
  const [content, setContent] = useState('');
  const [validationError, setValidationError] = useState<string | null>(null);

  // When data arrives, populate form
  useEffect(() => {
    if (data?.content) {
      setContent(data.content);
    }
  }, [data]);

  // Mutation to publish
  const mutation = useMutation({
    mutationFn: () => publishEssay({ content, author: username }),
    onSuccess: () => {
	    queryClient.invalidateQueries({ queryKey: ['myEssay', username] });
    },
  });

  const handleSubmit = () => {
    const len = content.trim().length;
    if (len < 1024 || len > 4096) {
      setValidationError('Essay must be between 1024 and 4096 characters.');
      return;
    }
    setValidationError(null);
    mutation.mutate();
  };

  if (isLoading) return <CircularProgress />;

  return (
    <Box maxWidth="md" mx="auto">
      <Typography variant="h5" gutterBottom>
        {data ? 'Edit Your Essay' : 'Publish Your Essay'}
      </Typography>

      {mutation.isSuccess && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Essay published successfully!
        </Alert>
      )}
      {mutation.isError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {(mutation.error as any).response?.data?.error || 'Failed to publish.'}
        </Alert>
      )}
      {validationError && (
        <Alert severity="warning" sx={{ mb: 2 }}>
          {validationError}
        </Alert>
      )}

      <TextField
        label="Your Essay"
        multiline
        minRows={12}
        fullWidth
        value={content}
        onChange={e => setContent(e.target.value)}
        helperText={`Length: ${content.trim().length} characters (1024–4096)`}
        sx={{ mb: 2 }}
      />

      <Button
        variant="contained"
        disabled={mutation.status === 'pending'}
        onClick={handleSubmit}
      >
        {mutation.status === 'pending'
          ? 'Publishing…'
          : data
          ? 'Update Essay'
          : 'Publish Essay'}
      </Button>
    </Box>
  );
};

export default MyEssay;
