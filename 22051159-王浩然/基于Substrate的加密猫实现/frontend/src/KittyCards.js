import ExtrinsicUnknown from '@polkadot/types/extrinsic/ExtrinsicUnknown';
import React from 'react';
import { Button, Card, Grid, Message, Modal, Form, Label } from 'semantic-ui-react';

import KittyAvatar from './KittyAvatar';
import { TxButton } from './substrate-lib/components';

// --- About Modal ---

const TransferModal = props => {
  const { kitty, accountPair, setStatus } = props;
  const [open, setOpen] = React.useState(false);
  const [formValue, setFormValue] = React.useState({});

  const formChange = key => (ev, el) => {
    /* TODO: 加代码 */
    setFormValue(prev => ({...prev, [key]: el.value}));
  };

  const confirmAndClose = (unsub) => {
    unsub();
    setOpen(false);
  };

  return <Modal onClose={() => setOpen(false)} onOpen={() => setOpen(true)} open={open}
    trigger={<Button basic color='blue'>转让</Button>}>
    <Modal.Header>毛孩转让</Modal.Header>
    <Modal.Content><Form>
      <Form.Input fluid label='毛孩 ID' readOnly value={kitty.id}/>
      <Form.Input fluid label='转让对象' placeholder='对方地址' onChange={formChange('target')}/>
    </Form></Modal.Content>
    <Modal.Actions>
      <Button basic color='grey' onClick={() => setOpen(false)}>取消</Button>
      <TxButton
        accountPair={accountPair} label='确认转让' type='SIGNED-TX' setStatus={setStatus}
        onClick={confirmAndClose}
        attrs={{
          palletRpc: 'kittiesModule',
          callable: 'transfer',
          inputParams: [formValue.target, kitty.id],
          paramFields: [true, true]
        }}
      />
    </Modal.Actions>
  </Modal>;
};

// --- About Kitty Card ---
function stringToByte(str) {
  const bytes = [];
  let len, c;
  len = str.length;
  for(var i = 0; i < len; i++) {
      c = str.charCodeAt(i);
      if(c >= 0x010000 && c <= 0x10FFFF) {
          bytes.push(((c >> 18) & 0x07) | 0xF0);
          bytes.push(((c >> 12) & 0x3F) | 0x80);
          bytes.push(((c >> 6) & 0x3F) | 0x80);
          bytes.push((c & 0x3F) | 0x80);
      } else if(c >= 0x000800 && c <= 0x00FFFF) {
          bytes.push(((c >> 12) & 0x0F) | 0xE0);
          bytes.push(((c >> 6) & 0x3F) | 0x80);
          bytes.push((c & 0x3F) | 0x80);
      } else if(c >= 0x000080 && c <= 0x0007FF) {
          bytes.push(((c >> 6) & 0x1F) | 0xC0);
          bytes.push((c & 0x3F) | 0x80);
      } else {
          bytes.push(c & 0xFF);
      }
  }
  return bytes;
}

const KittyCard = props => {
  /*
    TODO: 加代码。这里会 UI 显示一张 `KittyCard` 是怎么样的。这里会用到：
    ```
    <KittyAvatar dna={dna} /> - 来描绘一只猫咪
    <TransferModal kitty={kitty} accountPair={accountPair} setStatus={setStatus}/> - 来作转让的弹出层
    ```
  */
  const { kitty, owner, price, accountPair, setStatus } = props;
  const { dna } = kitty;
  const dna_str = dna.toString();
  const dna_arr = stringToByte(dna_str.slice(2, dna_str.length));
  const owner_str = "" + owner;

  let is_owner = accountPair.address === owner_str;

  return (
    <Grid.Column width={4}>
      <Card.Content>
        <Card.Header>{kitty.id}</Card.Header>
        <KittyAvatar dna={dna_arr} />
        {/* 只有属于自己的kitty才能转让给其它人 */}
        {
          is_owner ? <TransferModal kitty={kitty} is_owner={is_owner} accountPair={accountPair} setStatus={setStatus}/> : null
        }
        <Card.Description>
          {1}
        </Card.Description>
      </Card.Content>
    </Grid.Column>
  );
};

const KittyCards = props => {
  const { kitties, accountPair,kittyOwners, kittyPrices,setStatus } = props;

  /* TODO: 加代码。这里会枚举所有的 `KittyCard` */
  return (
    <Grid>
      {kitties.map((kitty, index) => <KittyCard key={index} kitty={kitty} owner={kittyOwners[index]} price={kittyPrices[index]} accountPair={accountPair} setStatus={setStatus}/>)}
    </Grid>
  );
};

export default KittyCards;
