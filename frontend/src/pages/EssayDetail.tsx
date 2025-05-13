// src/pages/EssayDetail.tsx
import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../api';
import { getAccessToken } from '../auth';
import {
  Box, Typography, Divider, List, ListItem, ListItemText,
  TextField, Button, Alert, MenuItem, Select,
} from '@mui/material';
import type { SelectChangeEvent } from '@mui/material/Select';

interface Review {
  id: number;
  rank: number;
  content: string;
  author: string;
  created_at: string;
}

interface EssayWithReviews {
  id: number;
  author: string;
  content: string;
  created_at: string;
  reviews: Review[];
}

const fetchEssay = (author: string) =>
  api.get<EssayWithReviews>(`/essay/${author}`).then(r => r.data);

const submitReview = (req: {
  essayId: number;
  rank: number;
  content: string;
  author: string;
}) => api.post('/review', req).then(r => r.data);

const EssayDetail: React.FC = () => {
  const { author } = useParams<{ author: string }>();
  const qc = useQueryClient();
  const token = getAccessToken();

  const { data, isLoading, error } = useQuery<EssayWithReviews, Error>({
    queryKey: ['essay', author],
    queryFn: () => fetchEssay(author!),
    retry: false,
  });
  // Determine status message
  let statusMessage: string;
  if (!data || data.reviews.length === 0) {
    statusMessage = 'У этого эссе нет рецензий, поставьте как можно больший рейтинг';
  } else {
    const bestRank = Math.min(...data.reviews.map(r => r.rank));
    if (bestRank === 3) {
      statusMessage =
        'Есть рецензия ранга 3, вы можете помочь автору оценив эссе выше (1 или 2)';
    } else if (bestRank === 2) {
      statusMessage =
        'Есть рецензия ранга 2, вы можете помочь автору оценив эссе выше (1)';
    } else {
      statusMessage =
        'Есть рецензия ранга 1, ОТДАЙТЕ СВОЙ ГОЛОС ДРУГОМУ';
    }
  }
    // Review form state
  const [rank, setRank] = useState('1');
  const [content, setContent] = useState('');
  const [formError, setFormError] = useState<string|null>(null);

  const mutation = useMutation<any, Error, void>({
    mutationFn: () =>
      submitReview({
        essayId: data!.id,
        rank: Number(rank),
        content: content.trim(),
        author: localStorage.getItem('username')!,
      }),
    onSuccess: () => {
	    qc.invalidateQueries({ queryKey: ['essay', author] });
      setContent('');
      setRank('1');
    },
  });

  const handleReviewSubmit = () => {
    if (content.trim().length < 256 || content.trim().length > 1024) {
      setFormError('Review content length must be 256-1024 characters.');
      return;
    }
    if (Number(rank) < 1 || Number(rank) > 3) {
      setFormError('Rank must be between 1 and 3.');
      return;
    }
    setFormError(null);
    mutation.mutate();
  };

  if (isLoading) return <Typography>Loading…</Typography>;
  if (error || !data) return <Typography color="error">Error loading essay</Typography>;

  return (
    <Box>
    	{/* Status banner */}
      <Alert severity="info" sx={{ mb: 2 }}>
        {statusMessage}
      </Alert>
      {/* existing essay display */}
      <Typography variant="h5" gutterBottom>
        Essay by {data.author}
      </Typography>
      <Typography paragraph>{data.content}</Typography>
      <Divider sx={{ my: 2 }} />
      <Typography variant="h6">Reviews</Typography>
      {data.reviews.length === 0 && <Typography>No reviews yet.</Typography>}
      <List>
        {data.reviews.map(r => (
          <ListItem key={r.id} alignItems="flex-start">
            <ListItemText
              primary={`[${r.rank}/3] by ${r.author}`}
              secondary={r.content}
            />
          </ListItem>
        ))}
      </List>

      {/* write-review area */}
      <Divider sx={{ my: 3 }} />
      {token ? (
        <Box>
          <Typography variant="h6" gutterBottom>Write a Review</Typography>
          {formError && <Alert severity="warning" sx={{ mb: 2 }}>{formError}</Alert>}
          {mutation.isError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {(mutation.error as any).response?.data?.error || 'Failed to submit review.'}
            </Alert>
          )}
          {mutation.isSuccess && (
            <Alert severity="success" sx={{ mb: 2 }}>
              Review submitted!
            </Alert>
          )}

          <Typography variant="h6" gutterBottom>Essay rank</Typography>
          <Select
            value={rank}
            onChange={(e: SelectChangeEvent) => setRank(e.target.value)}
            sx={{ mb: 2, width: 120 }}
          >
            <MenuItem value="1">1</MenuItem>
            <MenuItem value="2">2</MenuItem>
            <MenuItem value="3">3</MenuItem>
          </Select>
          <TextField
            label="Your review"
            multiline
            minRows={4}
            fullWidth
            value={content}
            onChange={e => setContent(e.target.value)}
            helperText={`Length: ${content.trim().length} characters (256–1024)`}
            sx={{ mb: 2 }}
          />
          <Button
            variant="contained"
            disabled={mutation.status === 'pending'}
            onClick={handleReviewSubmit}
          >
            {mutation.status === 'pending' ? 'Submitting…' : 'Submit Review'}
          </Button>
        </Box>
      ) : (
        <Alert severity="info">
          Please <Typography component="span" color="primary">login</Typography> to write a review.
        </Alert>
      )}
    </Box>
  );
};

export default EssayDetail;
