import React, { useEffect, useState } from 'react';
import { Form, Grid } from 'semantic-ui-react';

import { useSubstrate } from './substrate-lib';
import { TxButton } from './substrate-lib/components';

import KittyCards from './KittyCards';

export default function Kitties (props) {
  const { api, keyring } = useSubstrate();
  const { accountPair } = props;

  const [kittyCnt, setKittyCnt] = useState(0);
  const [kittyDNAs, setKittyDNAs] = useState([]);
  const [kittyOwners, setKittyOwners] = useState([]);
  const [kittyPrices, setKittyPrices] = useState([]);
  const [kitties, setKitties] = useState([]);
  const [status, setStatus] = useState('');


  const getKitttIndex = () => {
    let kitty_index_args = [];
    for(let i = 0; i < kittyCnt; ++i){
      kitty_index_args.push(i);
    }
    return kitty_index_args;
  }

  const fetchKittyCnt = () => {
    /* TODO: 加代码，从 substrate 端读取数据过来 */
    let unsubscribe;
    api.query.kittiesModule.kittiesCount(cnt => {
      console.log("aaa");
      setKittyCnt(cnt.toNumber());
    }).then(unsub => {
      unsubscribe = unsub;
    }).catch(console.error);
    return () => unsubscribe && unsubscribe();
  };

  const fetchKitties = () => {
    /* TODO: 加代码，从 substrate 端读取数据过来 */
    let unsubscribe;
    let kitty_index_args = getKitttIndex();
    console.log("kitty_index_args: {}", kitty_index_args);
    api.query.kittiesModule.kitties.multi(kitty_index_args, (kitties) => {
      console.log("kitties: {}", kitties);
      let new_kitties = kitties.map((dna, i) => {
        return {dna: dna, id: i, is_owner: true};
      })
      setKitties(new_kitties);
    }).then(unsub => {
      unsubscribe = unsub;
    }).catch(console.error);
    return () => {return unsubscribe && unsubscribe()};
  };

  const fetchKittyOwner = () => {
    let unsubscribe;
    let kitty_index_args = getKitttIndex();
    api.query.kittiesModule.kittyOwners.multi(kitty_index_args, (kittyOwners) => {
      setKittyOwners(kittyOwners);
    }).then(u => {
      unsubscribe = u;
    }).catch(console.error)
    return () => {return unsubscribe && unsubscribe()};
  }
  const fethKittyPrice = () => {
    let unsubscribe;
    let kitty_index_args = getKitttIndex();
    api.query.kittiesModule.kittyOwners.multi(kitty_index_args, (kittyPrices) => {
      setKittyOwners(kittyPrices);
    }).then(u => {
      unsubscribe = u;
    }).catch(console.error);
    return () => {return unsubscribe && unsubscribe()};
  }

  const populateKitties = () => {
    /* TODO: 加代码，从 substrate 端读取数据过来 */
  };

  useEffect(fetchKittyCnt, [keyring, setKittyCnt, api.query.kittiesModule]);
  useEffect(fetchKitties, [kittyCnt, setKitties, api.query.kittiesModule]);
  useEffect(fetchKittyOwner, [kittyCnt, setKittyOwners, api.query.kittiesModule]);
  useEffect(fethKittyPrice, [kittyCnt, setKittyPrices, api.query.kittiesModule]);
  useEffect(populateKitties,  [kittyDNAs, kittyOwners]);


  return <Grid.Column width={16}>
    <h1>小毛孩</h1>
    <h2>现有共有 {kittyCnt} 只</h2>

    <KittyCards kitties= {kitties} kittyOwners={kittyOwners} kittyPrices={kittyPrices} accountPair={accountPair} setStatus={setStatus}/>

    <Form style={{ margin: '1em 0' }}>
      <Form.Field style={{ textAlign: 'center' }}>
        <TxButton
          accountPair={accountPair} label='创建小毛孩' type='SIGNED-TX' setStatus={setStatus}
          attrs={{
            palletRpc: 'kittiesModule',
            callable: 'create',
            inputParams: [],
            paramFields: []
          }}
        />
      </Form.Field>
    </Form>
    <div style={{ overflowWrap: 'break-word' }}>{status}</div>
  </Grid.Column>;
}
