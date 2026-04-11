const fs = require('fs');
let file = fs.readFileSync('frontend/src/routes/OrderDetailPage.tsx', 'utf8');

file = file.replace(
  'useAddOrderNote,',
  'useAddOrderNote,\n  useSplitOrder,\n  useMergeOrder,'
);

const splitDialog = `
function SplitDialog({ open, onClose, orderId, maxQuantity }: { open: boolean; onClose: () => void; orderId: string; maxQuantity: number }) {
  const [quantities, setQuantities] = useState('1,1');
  const splitMutation = useSplitOrder();
  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Split Order</DialogTitle>
      <DialogContent>
        <Typography variant="body2" gutterBottom>
          Enter comma-separated quantities. Total must equal {maxQuantity} or less.
        </Typography>
        <TextField autoFocus margin="dense" label="Quantities" type="text" fullWidth value={quantities} onChange={e => setQuantities(e.target.value)} />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={splitMutation.isPending}>Cancel</Button>
        <Button onClick={() => {
          const q = quantities.split(',').map(n => parseInt(n.trim(), 10)).filter(n => !isNaN(n));
          splitMutation.mutate({ id: orderId, quantities: q }, { onSuccess: onClose });
        }} disabled={splitMutation.isPending} variant="contained">Split</Button>
      </DialogActions>
    </Dialog>
  );
}

`;
file = file.replace('export function OrderDetailPage', splitDialog + 'export function OrderDetailPage');

file = file.replace('const [payOpen, setPayOpen] = useState(false);', 'const [payOpen, setPayOpen] = useState(false);\n  const [splitOpen, setSplitOpen] = useState(false);');

file = file.replace(
  '<Button variant="outlined" size="small" color="error" onClick={() => setCancelOpen(true)}>',
  '<Button variant="outlined" size="small" onClick={() => setSplitOpen(true)}>Split</Button>\n                <Button variant="outlined" size="small" color="error" onClick={() => setCancelOpen(true)}>'
);

file = file.replace(
  '{id && <PayDialog open={payOpen} onClose={() => setPayOpen(false)} orderId={id} />}',
  '{id && <PayDialog open={payOpen} onClose={() => setPayOpen(false)} orderId={id} />}\n        {id && order && <SplitDialog open={splitOpen} onClose={() => setSplitOpen(false)} orderId={id} maxQuantity={order.quantity} />}'
);

fs.writeFileSync('frontend/src/routes/OrderDetailPage.tsx', file);
